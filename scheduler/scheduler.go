package scheduler

import (
	"fmt"
	"sync"
	"time"

	"github.com/raintank/raintank-metric/schema"
	"github.com/raintank/raintank-probe/checks"
	"github.com/raintank/raintank-probe/probe"
	"github.com/raintank/raintank-probe/publisher"
	"github.com/raintank/worldping-api/pkg/log"
	m "github.com/raintank/worldping-api/pkg/models"
)

type RaintankProbeCheck interface {
	Run() (checks.CheckResult, error)
}

type CheckInstance struct {
	Ticker      *time.Ticker
	Exec        RaintankProbeCheck
	Check       *m.MonitorDTO
	State       m.CheckEvalResult
	StateChange time.Time
	LastError   string
	sync.Mutex
}

func NewCheckInstance(c *m.MonitorDTO) (*CheckInstance, error) {
	log.Info("Creating new CheckInstance for %s check for %s", c.MonitorTypeName, c.EndpointSlug)
	executor, err := GetCheck(c.MonitorTypeId, m.MonitorSettingsDTO(c.Settings).ToV2Setting(m.CheckType(c.MonitorTypeName)))
	if err != nil {
		return nil, err
	}
	instance := &CheckInstance{
		Check: c,
		Exec:  executor,
		State: m.EvalResultUnknown,
	}
	go instance.Run()
	return instance, nil
}

func (i *CheckInstance) Update(c *m.MonitorDTO) error {
	executor, err := GetCheck(c.MonitorTypeId, m.MonitorSettingsDTO(c.Settings).ToV2Setting(m.CheckType(c.MonitorTypeName)))
	if err != nil {
		return err
	}
	i.Lock()
	i.Ticker.Stop()
	i.Check = c
	i.Exec = executor
	i.Unlock()
	go i.Run()
	return nil
}

func (i *CheckInstance) Stop() {
	if i.Ticker != nil {
		i.Ticker.Stop()
	}
}

func (c *CheckInstance) Run() {
	c.Lock()
	log.Info("Starting execution loop for %s check for %s, Frequency: %d, Offset: %d", c.Check.MonitorTypeName, c.Check.EndpointSlug, c.Check.Frequency, c.Check.Offset)
	now := time.Now().Unix()
	waitTime := ((c.Check.Frequency + c.Check.Offset) - (now % c.Check.Frequency)) % c.Check.Frequency
	if waitTime == c.Check.Offset {
		waitTime = 0
	}
	log.Debug("executing %s check for %s in %d seconds", c.Check.MonitorTypeName, c.Check.EndpointSlug, waitTime)
	if waitTime > 0 {
		time.Sleep(time.Second * time.Duration(waitTime))
	}
	c.Ticker = time.NewTicker(time.Duration(c.Check.Frequency) * time.Second)
	c.Unlock()
	c.run(time.Now())
	for t := range c.Ticker.C {
		c.run(t)
	}
}

func (c *CheckInstance) run(t time.Time) {
	desc := fmt.Sprintf("%s check for %s", c.Check.MonitorTypeName, c.Check.EndpointSlug)
	log.Debug("Running %s", desc)
	results, err := c.Exec.Run()
	var metrics []*schema.MetricData
	if err != nil {
		log.Error(3, "Failed to execute %s", desc, err)
		return
	} else {
		metrics = results.Metrics(t, c.Check)
		log.Debug("got %d metrics for %s", len(metrics), desc)
		// check if we need to send any events.  Events are sent on state change, or if the error reason has changed
		// or the check has been in an error state for 10minutes.
		newState := m.EvalResultOK
		if msg := results.ErrorMsg(); msg != "" {
			log.Debug("%s failed: %s", desc, msg)
			newState = m.EvalResultCrit
			if (c.State != newState) || (msg != c.LastError) || (time.Since(c.StateChange) > time.Minute*10) {
				c.State = newState
				c.LastError = msg
				c.StateChange = time.Now()
				//send Error event.
				log.Info("%s is in error state", desc)
				event := schema.ProbeEvent{
					EventType: "monitor_state",
					OrgId:     c.Check.OrgId,
					Severity:  "ERROR",
					Source:    "monitor_collector",
					Timestamp: t.UnixNano() / int64(time.Millisecond),
					Message:   msg,
					Tags: map[string]string{
						"endpoint":     c.Check.EndpointSlug,
						"collector":    probe.Self.Slug,
						"monitor_type": c.Check.MonitorTypeName,
					},
				}
				publisher.Publisher.AddEvent(&event)
			}
		} else if c.State != newState {
			c.State = newState
			c.StateChange = time.Now()
			//send OK event.
			log.Info("%s is now in OK state", desc)
			event := schema.ProbeEvent{
				EventType: "monitor_state",
				OrgId:     c.Check.OrgId,
				Severity:  "OK",
				Source:    "monitor_collector",
				Timestamp: t.UnixNano() / int64(time.Millisecond),
				Message:   "Monitor now Ok.",
				Tags: map[string]string{
					"endpoint":     c.Check.EndpointSlug,
					"collector":    probe.Self.Slug,
					"monitor_type": c.Check.MonitorTypeName,
				},
			}
			publisher.Publisher.AddEvent(&event)
		}
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
		OrgId:      int(c.Check.OrgId),
		Name:       fmt.Sprintf("litmus.%s.%s.%s.ok_state", c.Check.EndpointSlug, probe.Self.Slug, c.Check.MonitorTypeName),
		Metric:     fmt.Sprintf("litmus.%s.ok_state", c.Check.MonitorTypeName),
		Interval:   int(c.Check.Frequency),
		Unit:       "state",
		TargetType: "gauge",
		Time:       t.Unix(),
		Tags: []string{
			fmt.Sprintf("endpoint:%s", c.Check.EndpointSlug),
			fmt.Sprintf("probe:%s", probe.Self.Slug),
			fmt.Sprintf("checkType:%s", c.Check.MonitorTypeName),
		},
		Value: okState,
	}, &schema.MetricData{
		OrgId:      int(c.Check.OrgId),
		Name:       fmt.Sprintf("litmus.%s.%s.%s.error_state", c.Check.EndpointSlug, probe.Self.Slug, c.Check.MonitorTypeName),
		Metric:     fmt.Sprintf("litmus.%s.error_state", c.Check.MonitorTypeName),
		Interval:   int(c.Check.Frequency),
		Unit:       "state",
		TargetType: "gauge",
		Time:       t.Unix(),
		Tags: []string{
			fmt.Sprintf("endpoint:%s", c.Check.EndpointSlug),
			fmt.Sprintf("probe:%s", probe.Self.Slug),
			fmt.Sprintf("checkType:%s", c.Check.MonitorTypeName),
		},
		Value: errState,
	})

	for _, m := range metrics {
		m.SetId()
	}

	//publish metrics to TSDB
	publisher.Publisher.Add(metrics)
}

type Scheduler struct {
	sync.RWMutex
	Checks map[int64]*CheckInstance
}

func New() *Scheduler {
	return &Scheduler{
		Checks: make(map[int64]*CheckInstance),
	}
}

func (s *Scheduler) Close() {
	s.Lock()
	for _, instance := range s.Checks {
		instance.Stop()
	}
	s.Checks = make(map[int64]*CheckInstance)
	return
}

func (s *Scheduler) Refresh(checks []*m.MonitorDTO) {
	log.Debug("refreshing checks, there are %d", len(checks))
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
				err := existing.Update(c)
				if err != nil {
					log.Error(3, "Unable to update check instance for checkId=%d", c.Id, err)
					existing.Stop()
					delete(s.Checks, c.Id)
				}
			}
		} else {
			log.Debug("new check definition found for checkId=%d.", c.Id)
			instance, err := NewCheckInstance(c)
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
			instance.Stop()
			delete(s.Checks, id)
		}
	}
	s.Unlock()
	log.Debug("refresh complete")
	return
}

func (s *Scheduler) Create(check *m.MonitorDTO) {
	log.Debug("creating %s check for %s", check.MonitorTypeName, check.EndpointSlug)

	return
}

func (s *Scheduler) Update(check *m.MonitorDTO) {
	log.Debug("updating %s check for %s", check.MonitorTypeName, check.EndpointSlug)
	return
}

func (s *Scheduler) Remove(check *m.MonitorDTO) {
	log.Debug("removing %s check for %s", check.MonitorTypeName, check.EndpointSlug)
	return
}

func GetCheck(checkType int64, settings map[string]interface{}) (RaintankProbeCheck, error) {
	switch checkType {
	case 3:
		return checks.NewRaintankPingProbe(settings)
	case 4:
		return checks.NewRaintankDnsProbe(settings)
	case 1:
		return checks.NewRaintankHTTPProbe(settings)
	case 2:
		return checks.NewRaintankHTTPSProbe(settings)
	default:
		return nil, fmt.Errorf("unknown check type %d ", checkType)
	}

}
