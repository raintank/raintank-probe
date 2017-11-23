package publisher

import (
	"bytes"
	"crypto/tls"
	"hash/fnv"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/golang/snappy"
	eventMsg "github.com/grafana/worldping-gw/msg"
	"github.com/jpillora/backoff"
	"github.com/raintank/worldping-api/pkg/log"
	"gopkg.in/raintank/schema.v1"
	"gopkg.in/raintank/schema.v1/msg"
)

var (
	Publisher          *Tsdb
	maxMetricsPerFlush = 10000
	maxEventsPerFlush  = 10000
	maxFlushWait       = time.Millisecond * 500
)

func Init(u *url.URL, apiKey string, concurrency int) {
	Publisher = NewTsdb(u, apiKey, concurrency)
}

func Stop() {
	Publisher.Stop()
}

type Tsdb struct {
	sync.Mutex
	concurrency        int
	tsdbUrl            string
	tsdbKey            string
	metricsWriteQueues []chan []byte
	eventsWriteQueue   chan []byte
	shutdown           chan struct{}
	wg                 *sync.WaitGroup
	metricsIn          chan *schema.MetricData
	eventsIn           chan *eventMsg.ProbeEvent
	client             *http.Client
}

func NewTsdb(u *url.URL, apiKey string, concurrency int) *Tsdb {
	tsdbUrl := strings.TrimSuffix(u.String(), "/")
	t := &Tsdb{
		tsdbUrl:            tsdbUrl,
		tsdbKey:            apiKey,
		concurrency:        concurrency,
		metricsWriteQueues: make([]chan []byte, concurrency),
		eventsWriteQueue:   make(chan []byte, concurrency),
		shutdown:           make(chan struct{}),
		metricsIn:          make(chan *schema.MetricData, 1000000),
		eventsIn:           make(chan *eventMsg.ProbeEvent, 50000),
		wg:                 &sync.WaitGroup{},
	}
	for i := 0; i < concurrency; i++ {
		t.metricsWriteQueues[i] = make(chan []byte, 100)
	}
	// start off with a transport the same as Go's DefaultTransport
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
	// disable http 2.0 because there seems to be a compatibility problem between nginx hosts and the golang http2 implementation
	// which would occasionally result in bogus `400 Bad Request` errors.
	transport.TLSNextProto = make(map[string]func(authority string, c *tls.Conn) http.RoundTripper)

	t.client = &http.Client{
		Timeout: time.Second * 10,
	}
	//t.client.Transport = transport
	go t.run()
	return t
}

// Add metrics to the input buffer
func (t *Tsdb) Add(metrics []*schema.MetricData) {
	for _, m := range metrics {
		t.metricsIn <- m
	}
}

func (t *Tsdb) AddEvent(event *eventMsg.ProbeEvent) {
	t.eventsIn <- event
}

func (t *Tsdb) run() {
	metrics := make([][]*schema.MetricData, t.concurrency)
	events := make([]*eventMsg.ProbeEvent, 0, maxEventsPerFlush)
	for i := 0; i < t.concurrency; i++ {
		// buffers for holding metrics before flushing.
		metrics[i] = make([]*schema.MetricData, 0, maxMetricsPerFlush)

		// start up our goroutines for writing metrics to tsdb-gw
		t.wg.Add(1)
		go t.flushMetrics(i)

		// start up our goroutines for writing events to tsdb-gw
		t.wg.Add(1)
		go t.flushEvents()
	}

	flushMetrics := func(shard int) {
		if len(metrics[shard]) == 0 {
			return
		}
		mda := schema.MetricDataArray(metrics[shard])
		data, err := msg.CreateMsg(mda, 0, msg.FormatMetricDataArrayMsgp)
		if err != nil {
			panic(err)
		}
		t.metricsWriteQueues[shard] <- data
		metrics[shard] = metrics[shard][:0]
	}

	flushEvents := func() {
		if len(events) == 0 {
			return
		}
		data, err := eventMsg.CreateProbeEventsMsg(events)
		if err != nil {
			panic(err)
		}
		t.eventsWriteQueue <- data
		events = events[:0]
	}

	hasher := fnv.New32a()

	ticker := time.NewTicker(maxFlushWait)
	var buf []byte
	for {
		select {
		case md := <-t.metricsIn:
			//re-use our []byte slice to save an allocation.
			buf = md.KeyBySeries(buf[:0])
			hasher.Reset()
			hasher.Write(buf)
			shard := int(hasher.Sum32() % uint32(t.concurrency))
			metrics[shard] = append(metrics[shard], md)
			if len(metrics[shard]) == maxMetricsPerFlush {
				flushMetrics(shard)
			}
		case <-ticker.C:
			for shard := 0; shard < t.concurrency; shard++ {
				flushMetrics(shard)
			}
			flushEvents()
		case event := <-t.eventsIn:
			events = append(events, event)
			if len(events) == maxEventsPerFlush {
				flushEvents()
			}
		case <-t.shutdown:
			for shard := 0; shard < t.concurrency; shard++ {
				flushMetrics(shard)
				close(t.metricsWriteQueues[shard])
			}
			flushEvents()
			close(t.eventsWriteQueue)
			return
		}
	}
}

func (t *Tsdb) Stop() {
	close(t.shutdown)
	done := make(chan struct{})
	go func() {
		t.wg.Wait()
		close(done)
	}()
	select {
	case <-time.After(time.Minute):
		log.Info("timed out waiting for publisher to stop.")
		return
	case <-done:
		return
	}
}

func (t *Tsdb) flushMetrics(shard int) {
	q := t.metricsWriteQueues[shard]
	b := &backoff.Backoff{
		Min:    100 * time.Millisecond,
		Max:    time.Minute,
		Factor: 1.5,
		Jitter: true,
	}
	body := new(bytes.Buffer)
	var bodyLen int
	defer t.wg.Done()
	for data := range q {
		for {
			pre := time.Now()
			body.Reset()
			snappyBody := snappy.NewWriter(body)
			snappyBody.Write(data)
			bodyLen = body.Len()
			req, err := http.NewRequest("POST", t.tsdbUrl+"/metrics", body)
			if err != nil {
				panic(err)
			}
			req.Header.Add("Authorization", "Bearer "+t.tsdbKey)
			req.Header.Add("Content-Type", "rt-metric-binary-snappy")
			resp, err := t.client.Do(req)
			diff := time.Since(pre)
			if err == nil && resp.StatusCode >= 200 && resp.StatusCode < 300 {
				b.Reset()
				log.Debug("GrafanaNet sent metrics in %s -msg size %d", diff, bodyLen)
				resp.Body.Close()
				ioutil.ReadAll(resp.Body)
				break
			}
			dur := b.Duration()
			if err != nil {
				log.Warn("GrafanaNet failed to submit metrics: %s will try again in %s (this attempt took %s)", err, dur, diff)
			} else {
				buf := make([]byte, 300)
				n, _ := resp.Body.Read(buf)
				log.Warn("GrafanaNet failed to submit metrics: http %d - %s will try again in %s (this attempt took %s)", resp.StatusCode, buf[:n], dur, diff)
				resp.Body.Close()
			}

			time.Sleep(dur)
		}
	}
}

func (t *Tsdb) flushEvents() {
	q := t.eventsWriteQueue
	b := &backoff.Backoff{
		Min:    100 * time.Millisecond,
		Max:    time.Minute,
		Factor: 1.5,
		Jitter: true,
	}
	body := new(bytes.Buffer)
	var bodyLen int
	defer t.wg.Done()
	for data := range q {
		for {
			pre := time.Now()
			body.Reset()
			snappyBody := snappy.NewWriter(body)
			snappyBody.Write(data)
			bodyLen = body.Len()
			req, err := http.NewRequest("POST", t.tsdbUrl+"/events", body)
			if err != nil {
				panic(err)
			}
			req.Header.Add("Authorization", "Bearer "+t.tsdbKey)
			req.Header.Add("Content-Type", "rt-metric-binary-snappy")
			resp, err := t.client.Do(req)
			diff := time.Since(pre)
			if err == nil && resp.StatusCode >= 200 && resp.StatusCode < 300 {
				b.Reset()
				log.Debug("GrafanaNet sent event in %s -msg size %d", diff, bodyLen)
				resp.Body.Close()
				ioutil.ReadAll(resp.Body)
				break
			}
			dur := b.Duration()
			if err != nil {
				log.Warn("GrafanaNet failed to submit event: %s will try again in %s (this attempt took %s)", err, dur, diff)
			} else {
				buf := make([]byte, 300)
				n, _ := resp.Body.Read(buf)
				log.Warn("GrafanaNet failed to submit event: http %d - %s will try again in %s (this attempt took %s)", resp.StatusCode, buf[:n], dur, diff)
				resp.Body.Close()
			}

			time.Sleep(dur)
		}
	}
}
