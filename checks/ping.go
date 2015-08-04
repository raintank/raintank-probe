package checks

import (
	"encoding/json"
	"fmt"
	"github.com/tatsushid/go-fastping"
	"math"
	"net"
	"os"
	"sort"
	"time"
)

const count = 5

type PingResult struct {
	Loss  *float64 `json:"loss"`
	Min   *float64 `json:"min"`
	Max   *float64 `json:"max"`
	Avg   *float64 `json:"avg"`
	Mean  *float64 `json:"mean"`
	Mdev  *float64 `json:"mdev"`
	Error *string  `json:"error"`
}

type RaintankProbePing struct {
	Hostname string      `json:"hostname"`
	Result   *PingResult `json:"-"`
}

func NewRaintankPingProbe(body []byte) (*RaintankProbePing, error) {
	p := RaintankProbePing{}
	err := json.Unmarshal(body, &p)
	if err != nil {
		return nil, fmt.Errorf("failed to parse settings. " + err.Error())
	}
	return &p, nil
}

func (p *RaintankProbePing) Results() interface{} {
	return p.Result
}

func (p *RaintankProbePing) Run() error {
	p.Result = &PingResult{}

	var ipAddr string

	// get IP from hostname.
	addrs, err := net.LookupHost(p.Hostname)
	if err != nil || len(addrs) < 1 {
		msg := "failed to resolve hostname to IP."
		p.Result.Error = &msg
		return nil
	}
	ipAddr = addrs[0]

	pinger := fastping.NewPinger()
	results := make([]float64, 0)

	if err := pinger.AddIP(ipAddr); err != nil {
		msg := "failed to resolve hostname to IP."
		p.Result.Error = &msg
		return nil
	}

	//perform the Pings
	pinger.OnRecv = func(addr *net.IPAddr, rtt time.Duration) {
		//fmt.Printf("IP Addr: %s receive, RTT: %v\n", addr.String(), rtt)
		results = append(results, rtt.Seconds()*1000)
	}
	for i := 0; i < count; i++ {
		err := pinger.Run()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
	}

	// derive stats from results.
	successCount := len(results)
	failCount := float64(count - successCount)

	tsum := 0.0
	tsum2 := 0.0
	min := math.Inf(1)
	max := 0.0
	for _, r := range results {
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
		sort.Float64s(results)
		mean := results[successCount/2]
		p.Result.Mean = &mean
		p.Result.Min = &min
		p.Result.Max = &max
	}
	if failCount == 0 {
		loss := 0.0
		p.Result.Loss = &loss
	} else {
		loss := 100.0 * (failCount / float64(count))
		p.Result.Loss = &loss
	}
	if *p.Result.Loss == 100.0 {
		error := "100% packet loss"
		p.Result.Error = &error
	}
	return nil
}
