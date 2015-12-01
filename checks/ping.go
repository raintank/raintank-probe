package checks

import (
	"encoding/json"
	"fmt"
	"math"
	"net"
	"sort"
	"time"

	"github.com/raintank/go-pinger"
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
	Loss  *float64 `json:"loss"`
	Min   *float64 `json:"min"`
	Max   *float64 `json:"max"`
	Avg   *float64 `json:"avg"`
	Mean  *float64 `json:"mean"`
	Mdev  *float64 `json:"mdev"`
	Error *string  `json:"error"`
}

// Our check definition.
type RaintankProbePing struct {
	Hostname string      `json:"hostname"`
	Timeout  int         `json:"timeout"`
	Result   *PingResult `json:"-"`
}

// parse the json request body to build our check definition.
func NewRaintankPingProbe(body []byte) (*RaintankProbePing, error) {
	p := RaintankProbePing{}
	err := json.Unmarshal(body, &p)
	if err != nil {
		return nil, fmt.Errorf("failed to parse settings. " + err.Error())
	}
	return &p, nil
}

// return the results of the check
func (p *RaintankProbePing) Results() interface{} {
	return p.Result
}

// run the check. this is executed in a goroutine.
func (p *RaintankProbePing) Run() error {
	deadline := time.Now().Add(time.Second * time.Duration(p.Timeout))
	p.Result = &PingResult{}

	var ipAddr string

	// get IP from hostname.
	addrs, err := net.LookupHost(p.Hostname)
	if err != nil || len(addrs) < 1 {
		msg := "failed to resolve hostname to IP."
		p.Result.Error = &msg
		return nil
	}
	if time.Now().After(deadline) {
		msg := "timeout resolving IP address of hostname."
		p.Result.Error = &msg
		return nil
	}
	ipAddr = addrs[0]

	resultsChan, err := GlobalPinger.Ping(ipAddr, count, deadline)
	if err != nil {
		return err
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
		p.Result.Avg = &avg
		root := math.Sqrt((tsum2 / float64(successCount)) - ((tsum / float64(successCount)) * (tsum / float64(successCount))))
		p.Result.Mdev = &root
		sort.Float64s(measurements)
		mean := measurements[successCount/2]
		p.Result.Mean = &mean
		p.Result.Min = &min
		p.Result.Max = &max
	}
	if failCount == 0 {
		loss := 0.0
		p.Result.Loss = &loss
	} else {
		loss := 100.0 * (float64(failCount) / float64(results.Sent))
		p.Result.Loss = &loss
	}
	if *p.Result.Loss == 100.0 {
		error := "100% packet loss"
		p.Result.Error = &error
	}

	return nil
}
