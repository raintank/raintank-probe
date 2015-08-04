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

type PingSettings struct {
	Hostname string `json:"hostname"`
}

func Ping(body []byte) (*PingResult, error) {
	settings := PingSettings{}
	err := json.Unmarshal(body, &settings)
	if err != nil {
		return nil, fmt.Errorf("failed to parse settings. " + err.Error())
	}

	result := PingResult{}

	var ipAddr string

	// get IP from hostname.
	addrs, err := net.LookupHost(settings.Hostname)
	if err != nil || len(addrs) < 1 {
		msg := "failed to resolve hostname to IP."
		result.Error = &msg
		return &result, nil
	}
	ipAddr = addrs[0]

	p := fastping.NewPinger()
	results := make([]float64, 0)

	if err := p.AddIP(ipAddr); err != nil {
		msg := "failed to resolve hostname to IP."
		result.Error = &msg
		return &result, nil
	}

	//perform the Pings
	p.OnRecv = func(addr *net.IPAddr, rtt time.Duration) {
		//fmt.Printf("IP Addr: %s receive, RTT: %v\n", addr.String(), rtt)
		results = append(results, rtt.Seconds()*1000)
	}
	for i := 0; i < count; i++ {
		err := p.Run()
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
	successfullResults := make([]float64, 0)
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
		result.Avg = &avg
		root := math.Sqrt((tsum2 / float64(successCount)) - ((tsum / float64(successCount)) * (tsum / float64(successCount))))
		result.Mdev = &root
		sort.Float64s(successfullResults)
		mean := successfullResults[successCount/2]
		result.Mean = &mean
		result.Min = &min
		result.Max = &max
	}
	if failCount == 0 {
		loss := 0.0
		result.Loss = &loss
	} else {
		loss := 100.0 * (failCount / float64(count))
		result.Loss = &loss
	}
	if *result.Loss == 100.0 {
		error := "100% packet loss"
		result.Error = &error
	}
	return &result, nil
}
