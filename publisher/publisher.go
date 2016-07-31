package publisher

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/golang/snappy"
	"github.com/raintank/worldping-api/pkg/log"
	"gopkg.in/raintank/schema.v1"
	"gopkg.in/raintank/schema.v1/msg"
)

var (
	Publisher          *Tsdb
	maxMetricsPerFlush = 10000
)

type tsdbData struct {
	Path string
	Body []byte
}

func Init(u *url.URL, apiKey string, concurrency int) {
	Publisher = NewTsdb(u, apiKey, concurrency)
}

type Tsdb struct {
	sync.Mutex
	Concurrency  int
	Url          *url.URL
	ApiKey       string
	Metrics      []*schema.MetricData
	Events       []*schema.ProbeEvent
	flushEvents  chan struct{}
	flushMetrics chan struct{}
	closeChan    chan struct{}
	dataChan     chan tsdbData
}

func NewTsdb(u *url.URL, apiKey string, concurrency int) *Tsdb {
	t := &Tsdb{
		Metrics:      make([]*schema.MetricData, 0),
		flushMetrics: make(chan struct{}),
		flushEvents:  make(chan struct{}),
		Events:       make([]*schema.ProbeEvent, 0),
		Url:          u,
		ApiKey:       apiKey,
		Concurrency:  concurrency,
		dataChan:     make(chan tsdbData, concurrency),
	}
	go t.Run()
	return t
}

func (t *Tsdb) Add(metrics []*schema.MetricData) {
	log.Debug("received %d new metrics", len(metrics))
	t.Lock()
	t.Metrics = append(t.Metrics, metrics...)
	numMetrics := len(t.Metrics)
	t.Unlock()
	if numMetrics > maxMetricsPerFlush {
		//non-blocking send on the channel. If there is already
		// an item in the channel we dont need to add another.
		select {
		default:
			log.Debug("flushMetrics channel blocked.")
		case t.flushMetrics <- struct{}{}:
		}
	}
}

func (t *Tsdb) AddEvent(event *schema.ProbeEvent) {
	t.Lock()
	t.Events = append(t.Events, event)
	t.Unlock()
	//non-blocking send on the channel. If there is already
	// an item in the channel we dont need to add another.
	select {
	default:
		log.Debug("flushEvents channel blocked.")
	case t.flushEvents <- struct{}{}:
	}
}

func (t *Tsdb) Flush() {
	t.Lock()
	if len(t.Metrics) == 0 {
		t.Unlock()
		return
	}
	metrics := make([]*schema.MetricData, len(t.Metrics))
	copy(metrics, t.Metrics)
	t.Metrics = t.Metrics[:0]
	t.Unlock()

	// Write the metrics to our HTTP server.
	log.Debug("writing %d metrics to API", len(metrics))
	batches := schema.Reslice(metrics, maxMetricsPerFlush*2)
	for _, batch := range batches {
		id := time.Now().UnixNano()
		body, err := msg.CreateMsg(batch, id, msg.FormatMetricDataArrayMsgp)
		if err != nil {
			log.Error(3, "unable to convert metrics to MetricDataArrayMsgp.", "error", err)
			return
		}
		t.dataChan <- tsdbData{Path: "metrics", Body: body}
		log.Debug("%d metrics queud for delivery", len(batch))
	}
}

func (t *Tsdb) SendEvents() {
	t.Lock()
	if len(t.Events) == 0 {
		t.Unlock()
		return
	}
	events := make([]*schema.ProbeEvent, len(t.Events))
	copy(events, t.Events)
	t.Events = t.Events[:0]
	t.Unlock()
	for _, event := range events {
		id := time.Now().UnixNano()
		body, err := msg.CreateProbeEventMsg(event, id, msg.FormatProbeEventMsgp)
		if err != nil {
			log.Error(3, "Unable to convert event to ProbeEventMsgp.", "error", err)
			continue
		}
		t.dataChan <- tsdbData{Path: "events", Body: body}
	}
}

func (t *Tsdb) Run() {
	for i := 0; i < t.Concurrency; i++ {
		go t.sendData()
	}

	ticker := time.NewTicker(time.Second)
	last := time.Now()
	for {
		select {
		case <-ticker.C:
			if time.Since(last) >= time.Second {
				log.Debug("no flushes in last 1second. Flushing now.")
				last = time.Now()
				t.Flush()
				log.Debug("flush took %f seconds", time.Since(last).Seconds())
			}
		case <-t.flushMetrics:
			log.Debug("flush trigger received.")
			last = time.Now()
			t.Flush()
			log.Debug("flush took %f seconds", time.Since(last).Seconds())
		case <-t.flushEvents:
			t.SendEvents()
		case <-t.closeChan:
			close(t.dataChan)
			return
		}
	}
}
func (t *Tsdb) Close() {
	t.flushMetrics <- struct{}{}
	t.closeChan <- struct{}{}
}

func (t *Tsdb) sendData() {
	counter := 0
	bytesSent := 0
	last := time.Now()
	ticker := time.NewTicker(time.Second * 10)
	for {
		select {
		case <-ticker.C:
			if counter > 0 {
				log.Info("published %d (%d bytes) payloads in last %f seconds", counter, bytesSent, time.Since(last).Seconds())
				counter = 0
				bytesSent = 0
				last = time.Now()
			}
		case data := <-t.dataChan:
			u := t.Url.String() + data.Path
			body := new(bytes.Buffer)
			snappyBody := snappy.NewWriter(body)
			snappyBody.Write(data.Body)
			snappyBody.Close()
			req, err := http.NewRequest("POST", u, body)
			if err != nil {
				log.Error(3, "failed to create request payload. ", err)
				break
			}
			req.Header.Set("Content-Type", "rt-metric-binary-snappy")
			req.Header.Set("Authorization", "Bearer "+t.ApiKey)
			var reqBytesSent int
			sent := false
			for !sent {
				reqBytesSent = body.Len()
				if err := send(req); err != nil {
					log.Error(3, err.Error())
					time.Sleep(time.Second)
					body.Reset()
					snappyBody := snappy.NewWriter(body)
					snappyBody.Write(data.Body)
					snappyBody.Close()
				} else {
					sent = true
					log.Debug("sent %d bytes", reqBytesSent)
				}
			}
			bytesSent += reqBytesSent
			counter++
		}
	}
}

func send(req *http.Request) error {
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	respBody, _ := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("Posting data failed. %d - %s", resp.StatusCode, string(respBody))
	}
	return nil
}
