package scheduler

import (
	"fmt"
	"strings"
	"sync"
	"time"

	eventMsg "github.com/grafana/worldping-gw/msg"
	"github.com/raintank/raintank-probe/checks"
	"github.com/raintank/raintank-probe/probe"
	"github.com/raintank/raintank-probe/publisher"
	"github.com/raintank/worldping-api/pkg/log"
	m "github.com/raintank/worldping-api/pkg/models"
	"gopkg.in/raintank/schema.v1"
)

type RaintankProbeCheck interface {
	Run() (checks.CheckResult, error)
}

type CheckInstance struct {
	Ticker      *Ticker
	Exec        RaintankProbeCheck
	Check       *m.CheckWithSlug
	State       m.CheckEvalResult
	StateChange time.Time
	LastError   string
	stopped     bool
	sync.RWMutex
}

func NewCheckInstance(c *m.CheckWithSlug, probeHealthy bool) (*CheckInstance, error) {
	log.Info("Creating new CheckInstance for %s check for %s", c.Type, c.Slug)
	executor, err := GetCheck(c.Type, c.Settings)
	if err != nil {
		return nil, err
	}
	instance := &CheckInstance{
		Check:  c,
		Exec:   executor,
		State:  m.EvalResultUnknown,
		Ticker: NewTicker(c.Frequency, c.Offset),
	}
	go instance.loop()
	if probeHealthy {
		instance.Run()
	}
	return instance, nil
}

func (i *CheckInstance) Update(c *m.CheckWithSlug, probeHealthy bool) error {
	log.Info("updating execution thread of %s check for %s", c.Type, c.Slug)
	executor, err := GetCheck(c.Type, c.Settings)
	if err != nil {
		return err
	}
	i.Lock()
	i.Check = c
	i.Exec = executor
	i.Ticker.Update(c.Frequency, c.Offset)
	i.Unlock()
	return nil
}

func (i *CheckInstance) Stop() {
	i.RLock()
	log.Info("pausing execution thread of %s check for %s", i.Check.Type, i.Check.Slug)
	i.RUnlock()
	i.Ticker.Stop()
}

func (i *CheckInstance) Delete() {
	i.RLock()
	log.Info("stopping execution thread of %s check for %s", i.Check.Type, i.Check.Slug)
	i.RUnlock()
	i.Ticker.Delete()
}

func (i *CheckInstance) Run() {
	i.RLock()
	log.Info("enabling execution thread of %s check for %s", i.Check.Type, i.Check.Slug)
	i.RUnlock()
	i.Ticker.Start()
}

func (c *CheckInstance) loop() {
	c.RLock()
	log.Info("Starting execution loop for %s check for %s, Frequency: %d, Offset: %d", c.Check.Type, c.Check.Slug, c.Check.Frequency, c.Check.Offset)
	c.RUnlock()
	for t := range c.Ticker.C {
		c.run(t)
	}
	c.RLock()
	log.Info("execution loop for %s check for %s has ended.", c.Check.Type, c.Check.Slug)
	c.RUnlock()
}

func (c *CheckInstance) run(t time.Time) {
	c.Lock()
	desc := fmt.Sprintf("%s check for %s", c.Check.Type, c.Check.Slug)
	delta := time.Since(t)
	if delta > (100 * time.Millisecond) {
		log.Warn("%s check for %s is running late by %d milliseconds", c.Check.Type, c.Check.Slug, delta/time.Millisecond)
	}
	if (delta / time.Second) > time.Duration(c.Check.Frequency) {
		log.Error(3, "execution run of %s skipped due to being too old.", desc)
		c.Unlock()
		return
	}

	exec := c.Exec
	check := c.Check
	state := c.State
	stateChange := c.StateChange
	lastError := c.LastError
	c.Unlock()

	log.Debug("executing %s", desc)
	results, err := exec.Run()
	var metrics []*schema.MetricData
	if err != nil {
		log.Error(3, "Failed to execute %s", desc, err)
		return
	}
	metrics = results.Metrics(t, check)
	log.Debug("got %d metrics for %s", len(metrics), desc)
	// check if we need to send any events.  Events are sent on state change, or if the error reason has changed
	// or the check has been in an error state for 10minutes.
	newState := m.EvalResultOK
	if msg := results.ErrorMsg(); msg != "" {
		log.Debug("%s failed: %s", desc, msg)
		newState = m.EvalResultCrit
		if (state != newState) || (msg != lastError) || (time.Since(stateChange) > time.Minute*10) {
			c.Lock()
			c.State = newState
			c.LastError = msg
			c.StateChange = time.Now()
			c.Unlock()
			//send Error event.
			log.Info("%s is in error state", desc)
			event := eventMsg.ProbeEvent{
				EventType: "monitor_state",
				OrgId:     check.OrgId,
				Severity:  "ERROR",
				Source:    "monitor_collector",
				Timestamp: t.UnixNano() / int64(time.Millisecond),
				Message:   msg,
				Tags: map[string]string{
					"endpoint":     check.Slug,
					"collector":    probe.Self.Slug,
					"monitor_type": string(check.Type),
				},
			}
			publisher.Publisher.AddEvent(&event)
		}
	} else if state != newState {
		c.Lock()
		c.State = newState
		c.StateChange = time.Now()
		c.Unlock()
		//send OK event.
		log.Info("%s is now in OK state", desc)
		event := eventMsg.ProbeEvent{
			EventType: "monitor_state",
			OrgId:     check.OrgId,
			Severity:  "OK",
			Source:    "monitor_collector",
			Timestamp: t.UnixNano() / int64(time.Millisecond),
			Message:   "Monitor now Ok.",
			Tags: map[string]string{
				"endpoint":     check.Slug,
				"collector":    probe.Self.Slug,
				"monitor_type": string(check.Type),
			},
		}
		publisher.Publisher.AddEvent(&event)
	}

	// set or ok_state, error_state metrics.
	okState := 0.0
	errState := 0.0
	if c.State == m.EvalResultCrit {
		errState = 1
	} else {
		okState = 1
	}
	metrics = append(metrics, &schema.MetricData{
		OrgId:    int(check.OrgId),
		Name:     fmt.Sprintf("worldping.%s.%s.%s.ok_state", check.Slug, probe.Self.Slug, check.Type),
		Metric:   fmt.Sprintf("worldping.%s.%s.%s.ok_state", check.Slug, probe.Self.Slug, check.Type),
		Interval: int(check.Frequency),
		Unit:     "state",
		Mtype:    "gauge",
		Time:     t.Unix(),
		Tags:     nil,
		Value:    okState,
	}, &schema.MetricData{
		OrgId:    int(check.OrgId),
		Name:     fmt.Sprintf("worldping.%s.%s.%s.error_state", check.Slug, probe.Self.Slug, check.Type),
		Metric:   fmt.Sprintf("worldping.%s.%s.%s.error_state", check.Slug, probe.Self.Slug, check.Type),
		Interval: int(check.Frequency),
		Unit:     "state",
		Mtype:    "gauge",
		Time:     t.Unix(),
		Tags:     nil,
		Value:    errState,
	})

	for _, m := range metrics {
		m.SetId()
	}

	//publish metrics to TSDB
	publisher.Publisher.Add(metrics)
}

type Scheduler struct {
	sync.RWMutex
	Checks      map[int64]*CheckInstance
	HealthHosts []string
	Healthy     bool
}

func New(healthHosts string) *Scheduler {
	hosts := make([]string, 0)
	for _, h := range strings.Split(healthHosts, ",") {
		host := strings.TrimSpace(h)
		if host != "" {
			hosts = append(hosts, host)
		}
	}
	return &Scheduler{
		Checks:      make(map[int64]*CheckInstance),
		HealthHosts: hosts,
	}
}

func (s *Scheduler) IsHealthy() bool {
	s.RLock()
	healthy := s.Healthy
	s.RUnlock()
	return healthy
}

func (s *Scheduler) Close() {
	log.Info("Scheduler shutting down")
	s.Lock()
	for _, instance := range s.Checks {
		instance.Stop()
	}
	s.Checks = make(map[int64]*CheckInstance)
	log.Info("scheduler Shutdown complete.")
	return
}

func (s *Scheduler) Refresh(checks []*m.CheckWithSlug) {
	log.Info("refreshing checks, there are %d", len(checks))
	seenChecks := make(map[int64]struct{})
	s.Lock()
	for _, c := range checks {
		if !c.Enabled {
			continue
		}
		seenChecks[c.Id] = struct{}{}
		if existing, ok := s.Checks[c.Id]; ok {
			log.Debug("checkId=%d already running", c.Id)
			if c.Updated.After(existing.Check.Updated) {
				log.Info("syncing update to checkId=%d", c.Id)
				err := existing.Update(c, s.Healthy)
				if err != nil {
					log.Error(3, "Unable to update check instance for checkId=%d", c.Id, err)
					existing.Delete()
					delete(s.Checks, c.Id)
				}
			}
		} else {
			log.Debug("new check definition found for checkId=%d.", c.Id)
			instance, err := NewCheckInstance(c, s.Healthy)
			if err != nil {
				log.Error(3, "Unabled to create new check instance for checkId=%d.", c.Id, err)
			} else {
				s.Checks[c.Id] = instance
			}
		}
	}
	for id, instance := range s.Checks {
		if _, ok := seenChecks[id]; !ok {
			log.Info("checkId=%d no longer scheduled to this probe, removing it.", id)
			instance.Delete()
			delete(s.Checks, id)
		}
	}
	s.Unlock()
	log.Debug("refresh complete")
	return
}

func (s *Scheduler) Create(check *m.CheckWithSlug) {
	log.Info("creating %s check for %s", check.Type, check.Slug)
	s.Lock()
	if existing, ok := s.Checks[check.Id]; ok {
		log.Warn("recieved create event for check that is already running. checkId=%d", check.Id)
		existing.Delete()
		delete(s.Checks, check.Id)
	}
	instance, err := NewCheckInstance(check, s.Healthy)
	if err != nil {
		log.Error(3, "Unabled to create new check instance for checkId=%d.", check.Id, err)
	} else {
		s.Checks[check.Id] = instance
	}
	s.Unlock()
	return
}

func (s *Scheduler) Update(check *m.CheckWithSlug) {
	log.Info("updating %s check for %s", check.Type, check.Slug)
	s.Lock()
	if existing, ok := s.Checks[check.Id]; !ok {
		log.Warn("recieved update event for check that is not currently running. checkId=%d", check.Id)
		instance, err := NewCheckInstance(check, s.Healthy)
		if err != nil {
			log.Error(3, "Unabled to create new check instance for checkId=%d. %s", check.Id, err)
		} else {
			s.Checks[check.Id] = instance
		}

	} else {
		err := existing.Update(check, s.Healthy)
		if err != nil {
			log.Error(3, "Unable to update check instance for checkId=%d, %s", check.Id, err)
			existing.Delete()
			delete(s.Checks, check.Id)
		}
	}
	s.Unlock()
	return
}

func (s *Scheduler) Remove(check *m.CheckWithSlug) {
	log.Info("removing %s check for %s", check.Type, check.Slug)
	s.Lock()
	if existing, ok := s.Checks[check.Id]; !ok {
		log.Warn("recieved remove event for check that is not currently running. checkId=%d", check.Id)
	} else {
		existing.Delete()
		delete(s.Checks, check.Id)
	}
	s.Unlock()
	return
}

func GetCheck(checkType m.CheckType, settings map[string]interface{}) (RaintankProbeCheck, error) {
	switch checkType {
	case m.PING_CHECK:
		return checks.NewRaintankPingProbe(settings)
	case m.DNS_CHECK:
		return checks.NewRaintankDnsProbe(settings)
	case m.HTTP_CHECK:
		return checks.NewRaintankHTTPProbe(settings)
	case m.HTTPS_CHECK:
		return checks.NewRaintankHTTPSProbe(settings)
	default:
		return nil, fmt.Errorf("unknown check type %s ", checkType)
	}

}
