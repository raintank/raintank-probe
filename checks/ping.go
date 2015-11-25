package checks

import (
	"encoding/json"
	"fmt"
	"github.com/tatsushid/go-fastping"
	"math"
	"net"
	"os"
	"sort"
	"sync"
	"time"
)

// Number of pings to send to the host.
const count = 5

// Pinger object performs a Ping and returns
// the results to all registered listeners.
type Pinger struct {
	Address    string
	Results    []float64
	Started    bool
	Deadline   time.Time
	m          sync.Mutex
	resultChan []chan []float64
}

// Register a Listner to recieve results from a Ping job
func (p *Pinger) AddListener() <-chan []float64 {
	p.m.Lock()
	c := make(chan []float64)
	p.resultChan = append(p.resultChan, c)

	// if we haven't already started the ping, do so.
	if !p.Started {
		go p.Ping()
		p.Started = true
	}
	p.m.Unlock()
	return c
}

func (p *Pinger) Ping() {
	p.ping()
	// we lock here to prevent any new listeners being added.
	p.m.Lock()
	for _, c := range p.resultChan {
		c <- p.Results
		close(c)
	}
	p.resultChan = nil
	p.Started = false
	p.m.Unlock()
	// Any blocked AddListener calls, will now proceed and
	// run a fetch new ping results.
}

func (p *Pinger) ping() {
	fastpinger := fastping.NewPinger()
	// we hard code the timeout for each ping to 2seconds.
	fastpinger.MaxRTT = 2 * time.Second
	p.Results = make([]float64, 0)

	if err := fastpinger.AddIP(p.Address); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	//perform the Pings
	fastpinger.OnRecv = func(addr *net.IPAddr, rtt time.Duration) {
		//fmt.Printf("IP Addr: %s receive, RTT: %v\n", addr.String(), rtt)
		p.Results = append(p.Results, rtt.Seconds()*1000)
	}

	for i := 0; i < count; i++ {
		if time.Now().Before(p.Deadline) {
			err := fastpinger.Run()
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
			}
		}
	}
	return
}

type Job struct {
	Pinger *Pinger
	Count  int
}

type Coordinator struct {
	Jobs map[string]*Job
	m    sync.Mutex
}

func NewCoordinator() *Coordinator {
	return &Coordinator{
		Jobs: make(map[string]*Job),
	}
}

func (c *Coordinator) Ping(addr string, deadline time.Time) []float64 {
	c.m.Lock()
	job, ok := c.Jobs[addr]
	if !ok {
		//need to start new job.
		job = &Job{
			Pinger: &Pinger{Address: addr, Deadline: deadline},
			Count:  1,
		}
		c.Jobs[addr] = job
	} else {
		job.Count++
	}

	// add listener to job to get results.
	resultChan := job.Pinger.AddListener()
	c.m.Unlock()

	results := <-resultChan

	// we lock here to prevent anyone using this job while
	// we check if we are the last user of it.
	c.m.Lock()
	job.Count--
	if job.Count == 0 {
		delete(c.Jobs, addr)
	}
	c.m.Unlock()
	// Any calls that were blocked can now proceed. If
	// we deleted the job, they will create a new one.

	return results

}

// global co-oridinator shared between all go-routines.
var coordinator *Coordinator

func init() {
	coordinator = NewCoordinator()
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

	results := coordinator.Ping(ipAddr, deadline)

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
