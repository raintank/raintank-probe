package checks

import (
	"fmt"
	"math"
	"net"
	"sort"
	"time"

	"github.com/raintank/go-pinger"
	"github.com/raintank/raintank-probe/probe"
	m "github.com/raintank/worldping-api/pkg/models"
	"gopkg.in/raintank/schema.v1"
)

// Number of pings to send to the host.
const count = 5

// global co-oridinator shared between all go-routines.
var GlobalPinger *pinger.Pinger

func init() {
	var err error
	GlobalPinger, err = pinger.NewPinger("all", 10000)
	if err != nil {
		panic(err)
	}
	GlobalPinger.Start()
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

func (r *PingResult) Metrics(t time.Time, check *m.CheckWithSlug) []*schema.MetricData {
	metrics := make([]*schema.MetricData, 0)
	if r.Loss != nil {
		metrics = append(metrics, &schema.MetricData{
			OrgId:    int(check.OrgId),
			Name:     fmt.Sprintf("worldping.%s.%s.ping.loss", check.Slug, probe.Self.Slug),
			Metric:   "worldping.ping.loss",
			Interval: int(check.Frequency),
			Unit:     "percent",
			Mtype:    "gauge",
			Time:     t.Unix(),
			Tags: []string{
				fmt.Sprintf("endpoint:%s", check.Slug),
				fmt.Sprintf("monitor_type:%s", check.Type),
				fmt.Sprintf("probe:%s", probe.Self.Slug),
			},
			Value: *r.Loss,
		})
	}
	if r.Min != nil {
		metrics = append(metrics, &schema.MetricData{
			OrgId:    int(check.OrgId),
			Name:     fmt.Sprintf("worldping.%s.%s.ping.min", check.Slug, probe.Self.Slug),
			Metric:   "worldping.ping.min",
			Interval: int(check.Frequency),
			Unit:     "ms",
			Mtype:    "gauge",
			Time:     t.Unix(),
			Tags: []string{
				fmt.Sprintf("endpoint:%s", check.Slug),
				fmt.Sprintf("monitor_type:%s", check.Type),
				fmt.Sprintf("probe:%s", probe.Self.Slug),
			},
			Value: *r.Min,
		})
	}
	if r.Max != nil {
		metrics = append(metrics, &schema.MetricData{
			OrgId:    int(check.OrgId),
			Name:     fmt.Sprintf("worldping.%s.%s.ping.max", check.Slug, probe.Self.Slug),
			Metric:   "worldping.ping.max",
			Interval: int(check.Frequency),
			Unit:     "ms",
			Mtype:    "gauge",
			Time:     t.Unix(),
			Tags: []string{
				fmt.Sprintf("endpoint:%s", check.Slug),
				fmt.Sprintf("monitor_type:%s", check.Type),
				fmt.Sprintf("probe:%s", probe.Self.Slug),
			},
			Value: *r.Max,
		})
	}
	if r.Median != nil {
		metrics = append(metrics, &schema.MetricData{
			OrgId:    int(check.OrgId),
			Name:     fmt.Sprintf("worldping.%s.%s.ping.median", check.Slug, probe.Self.Slug),
			Metric:   "worldping.ping.median",
			Interval: int(check.Frequency),
			Unit:     "ms",
			Mtype:    "gauge",
			Time:     t.Unix(),
			Tags: []string{
				fmt.Sprintf("endpoint:%s", check.Slug),
				fmt.Sprintf("monitor_type:%s", check.Type),
				fmt.Sprintf("probe:%s", probe.Self.Slug),
			},
			Value: *r.Median,
		})
	}
	if r.Mdev != nil {
		metrics = append(metrics, &schema.MetricData{
			OrgId:    int(check.OrgId),
			Name:     fmt.Sprintf("worldping.%s.%s.ping.mdev", check.Slug, probe.Self.Slug),
			Metric:   "worldping.ping.mdev",
			Interval: int(check.Frequency),
			Unit:     "ms",
			Mtype:    "gauge",
			Time:     t.Unix(),
			Tags: []string{
				fmt.Sprintf("endpoint:%s", check.Slug),
				fmt.Sprintf("monitor_type:%s", check.Type),
				fmt.Sprintf("probe:%s", probe.Self.Slug),
			},
			Value: *r.Mdev,
		})
	}
	if r.Avg != nil {
		metrics = append(metrics, &schema.MetricData{
			OrgId:    int(check.OrgId),
			Name:     fmt.Sprintf("worldping.%s.%s.ping.mean", check.Slug, probe.Self.Slug),
			Metric:   "worldping.ping.mean",
			Interval: int(check.Frequency),
			Unit:     "ms",
			Mtype:    "gauge",
			Time:     t.Unix(),
			Tags: []string{
				fmt.Sprintf("endpoint:%s", check.Slug),
				fmt.Sprintf("monitor_type:%s", check.Type),
				fmt.Sprintf("probe:%s", probe.Self.Slug),
			},
			Value: *r.Avg,
		})
		metrics = append(metrics, &schema.MetricData{
			OrgId:    int(check.OrgId),
			Name:     fmt.Sprintf("worldping.%s.%s.ping.default", check.Slug, probe.Self.Slug),
			Metric:   "worldping.ping.default",
			Interval: int(check.Frequency),
			Unit:     "ms",
			Mtype:    "gauge",
			Time:     t.Unix(),
			Tags: []string{
				fmt.Sprintf("endpoint:%s", check.Slug),
				fmt.Sprintf("monitor_type:%s", check.Type),
				fmt.Sprintf("probe:%s", probe.Self.Slug),
			},
			Value: *r.Avg,
		})
	}

	return metrics
}

// Our check definition.
type RaintankProbePing struct {
	Hostname  string        `json:"hostname"`
	Timeout   time.Duration `json:"timeout"`
	IPVersion string        `json:"ipversion"`
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
	var t float64
	if !ok {
		t = 5.0
	} else {
		t, ok = timeout.(float64)
		if !ok {
			return nil, fmt.Errorf("invalid value for timeout, must be number.")
		}
	}
	if t <= 0.0 {
		return nil, fmt.Errorf("invalid value for timeout, must be greater then 0.")
	}
	p.Timeout = time.Duration(time.Millisecond * time.Duration(int(1000.0*t)))

	version, ok := settings["ipversion"]
	if !ok {
		p.IPVersion = "v4"
	} else {
		p.IPVersion, ok = version.(string)
		if !ok {
			return nil, fmt.Errorf("invalid value for ipversion, must be string.")
		}
	}
	if !(p.IPVersion == "v4" || p.IPVersion == "v6" || p.IPVersion == "any") {
		return nil, fmt.Errorf("ipversion must be v4, v6, or any.")
	}

	return &p, nil
}

func (p *RaintankProbePing) Run() (CheckResult, error) {
	deadline := time.Now().Add(p.Timeout)
	result := &PingResult{}

	// get IP from hostname.
	ipAddr, err := ResolveHost(p.Hostname, p.IPVersion)
	if err != nil {
		msg := err.Error()
		result.Error = &msg
		return result, nil
	}
	if time.Now().After(deadline) {
		msg := "timeout resolving IP address of hostname."
		result.Error = &msg
		return result, nil
	}

	// fmt.Printf("pinging %#v\n", ipAddr)
	results, err := GlobalPinger.Ping(net.ParseIP(ipAddr), count, p.Timeout)
	if err != nil {
		return nil, err
	}

	// derive stats from results.
	successCount := results.Received
	failCount := results.Sent - results.Received

	measurements := make([]float64, len(results.Latency))
	for i, m := range results.Latency {
		if m > p.Timeout {
			successCount--
			failCount++
			continue
		}
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

	return result, nil
}
