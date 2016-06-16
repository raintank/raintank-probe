package checks

import (
	"fmt"
	"math"
	"net"
	"sort"
	"time"

	"github.com/raintank/go-pinger"
	"github.com/raintank/raintank-metric/schema"
	"github.com/raintank/raintank-probe/probe"
	"github.com/raintank/worldping-api/pkg/log"
	m "github.com/raintank/worldping-api/pkg/models"
)

// Number of pings to send to the host.
const count = 5

// global co-oridinator shared between all go-routines.
var GlobalPinger *pinger.Pinger

func init() {
	GlobalPinger = pinger.NewPinger()
}

// results. we use pointers so that missing data will be
// encoded as 'null' in the json response.
type PingResult struct {
	Loss   *float64 `json:"loss"`
	Min    *float64 `json:"min"`
	Max    *float64 `json:"max"`
	Avg    *float64 `json:"avg"`
	Median *float64 `json:"median"`
	Mdev   *float64 `json:"mdev"`
	Error  *string  `json:"error"`
}

func (r *PingResult) ErrorMsg() string {
	if r.Error == nil {
		return ""
	}
	return *r.Error
}

func (r *PingResult) Metrics(t time.Time, check *m.MonitorDTO) []*schema.MetricData {
	metrics := make([]*schema.MetricData, 0)
	if r.Loss != nil {
		metrics = append(metrics, &schema.MetricData{
			OrgId:      int(check.OrgId),
			Name:       fmt.Sprintf("worldping.%s.%s.ping.loss", check.EndpointSlug, probe.Self.Slug),
			Metric:     "worldping.ping.loss",
			Interval:   int(check.Frequency),
			Unit:       "percent",
			TargetType: "gauge",
			Time:       t.Unix(),
			Tags: []string{
				fmt.Sprintf("endpoint:%s", check.EndpointSlug),
				fmt.Sprintf("probe:%s", probe.Self.Slug),
				"checkType:ping",
			},
			Value: *r.Loss,
		})
	}
	if r.Min != nil {
		metrics = append(metrics, &schema.MetricData{
			OrgId:      int(check.OrgId),
			Name:       fmt.Sprintf("worldping.%s.%s.ping.min", check.EndpointSlug, probe.Self.Slug),
			Metric:     "worldping.ping.min",
			Interval:   int(check.Frequency),
			Unit:       "ms",
			TargetType: "gauge",
			Time:       t.Unix(),
			Tags: []string{
				fmt.Sprintf("endpoint:%s", check.EndpointSlug),
				fmt.Sprintf("probe:%s", probe.Self.Slug),
				"checkType:ping",
			},
			Value: *r.Min,
		})
	}
	if r.Max != nil {
		metrics = append(metrics, &schema.MetricData{
			OrgId:      int(check.OrgId),
			Name:       fmt.Sprintf("worldping.%s.%s.ping.max", check.EndpointSlug, probe.Self.Slug),
			Metric:     "worldping.ping.max",
			Interval:   int(check.Frequency),
			Unit:       "ms",
			TargetType: "gauge",
			Time:       t.Unix(),
			Tags: []string{
				fmt.Sprintf("endpoint:%s", check.EndpointSlug),
				fmt.Sprintf("probe:%s", probe.Self.Slug),
				"checkType:ping",
			},
			Value: *r.Max,
		})
	}
	if r.Median != nil {
		metrics = append(metrics, &schema.MetricData{
			OrgId:      int(check.OrgId),
			Name:       fmt.Sprintf("worldping.%s.%s.ping.median", check.EndpointSlug, probe.Self.Slug),
			Metric:     "worldping.ping.median",
			Interval:   int(check.Frequency),
			Unit:       "ms",
			TargetType: "gauge",
			Time:       t.Unix(),
			Tags: []string{
				fmt.Sprintf("endpoint:%s", check.EndpointSlug),
				fmt.Sprintf("probe:%s", probe.Self.Slug),
				"checkType:ping",
			},
			Value: *r.Median,
		})
	}
	if r.Mdev != nil {
		metrics = append(metrics, &schema.MetricData{
			OrgId:      int(check.OrgId),
			Name:       fmt.Sprintf("worldping.%s.%s.ping.mdev", check.EndpointSlug, probe.Self.Slug),
			Metric:     "worldping.ping.mdev",
			Interval:   int(check.Frequency),
			Unit:       "percent",
			TargetType: "gauge",
			Time:       t.Unix(),
			Tags: []string{
				fmt.Sprintf("endpoint:%s", check.EndpointSlug),
				fmt.Sprintf("probe:%s", probe.Self.Slug),
				"checkType:ping",
			},
			Value: *r.Mdev,
		})
	}
	if r.Avg != nil {
		metrics = append(metrics, &schema.MetricData{
			OrgId:      int(check.OrgId),
			Name:       fmt.Sprintf("worldping.%s.%s.ping.avg", check.EndpointSlug, probe.Self.Slug),
			Metric:     "worldping.ping.avg",
			Interval:   int(check.Frequency),
			Unit:       "percent",
			TargetType: "gauge",
			Time:       t.Unix(),
			Tags: []string{
				fmt.Sprintf("endpoint:%s", check.EndpointSlug),
				fmt.Sprintf("probe:%s", probe.Self.Slug),
				"checkType:ping",
			},
			Value: *r.Avg,
		})
		metrics = append(metrics, &schema.MetricData{
			OrgId:      int(check.OrgId),
			Name:       fmt.Sprintf("worldping.%s.%s.ping.default", check.EndpointSlug, probe.Self.Slug),
			Metric:     "worldping.ping.default",
			Interval:   int(check.Frequency),
			Unit:       "percent",
			TargetType: "gauge",
			Time:       t.Unix(),
			Tags: []string{
				fmt.Sprintf("endpoint:%s", check.EndpointSlug),
				fmt.Sprintf("probe:%s", probe.Self.Slug),
				"checkType:ping",
			},
			Value: *r.Avg,
		})
	}

	return metrics
}

// Our check definition.
type RaintankProbePing struct {
	Hostname string        `json:"hostname"`
	Timeout  time.Duration `json:"timeout"`
}

// parse the json request body to build our check definition.
func NewRaintankPingProbe(settings map[string]interface{}) (*RaintankProbePing, error) {
	p := RaintankProbePing{}
	host, ok := settings["hostname"]
	if !ok {
		return nil, fmt.Errorf("no hostname passed.")
	}
	p.Hostname, ok = host.(string)
	if !ok {
		return nil, fmt.Errorf("invalid value for host, must be string.")
	}
	if p.Hostname == "" {
		return nil, fmt.Errorf("no host passed.")
	}

	timeout, ok := settings["timeout"]
	if !ok {
		return nil, fmt.Errorf("no timeout passed.")
	}
	t, ok := timeout.(float64)
	if !ok {
		return nil, fmt.Errorf("invalid value for timeout, must be number.")
	}
	if t <= 0.0 {
		return nil, fmt.Errorf("invalid value for timeout, must be greater then 0.")
	}
	p.Timeout = time.Duration(time.Millisecond * time.Duration(int(1000.0*t)))

	return &p, nil
}

func (p *RaintankProbePing) Run() (CheckResult, error) {
	deadline := time.Now().Add(p.Timeout)
	result := &PingResult{}

	var ipAddr string

	// get IP from hostname.
	addrs, err := net.LookupHost(p.Hostname)
	if err != nil || len(addrs) < 1 {
		msg := "failed to resolve hostname to IP."
		result.Error = &msg
		return result, nil
	}
	if time.Now().After(deadline) {
		msg := "timeout resolving IP address of hostname."
		result.Error = &msg
		return result, nil
	}
	ipAddr = addrs[0]

	resultsChan, err := GlobalPinger.Ping(ipAddr, count, deadline)
	if err != nil {
		return nil, err
	}

	results := <-resultsChan

	// derive stats from results.
	successCount := results.Received
	failCount := results.Sent - results.Received

	measurements := make([]float64, len(results.Latency))
	for i, m := range results.Latency {
		measurements[i] = m.Seconds() * 1000
	}

	tsum := 0.0
	tsum2 := 0.0
	min := math.Inf(1)
	max := 0.0
	for _, r := range measurements {
		if r > max {
			max = r
		}
		if r < min {
			min = r
		}
		tsum += r
		tsum2 += (r * r)
	}

	if successCount > 0 {
		avg := tsum / float64(successCount)
		result.Avg = &avg
		root := math.Sqrt((tsum2 / float64(successCount)) - ((tsum / float64(successCount)) * (tsum / float64(successCount))))
		result.Mdev = &root
		sort.Float64s(measurements)
		median := measurements[successCount/2]
		result.Median = &median
		result.Min = &min
		result.Max = &max
	}
	if failCount == 0 {
		loss := 0.0
		result.Loss = &loss
	} else {
		loss := 100.0 * (float64(failCount) / float64(results.Sent))
		result.Loss = &loss
	}
	if *result.Loss == 100.0 {
		errorMsg := "100% packet loss"
		result.Error = &errorMsg
	}

	log.Debug("Ping check completed.")

	return result, nil
}
