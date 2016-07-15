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
	"gopkg.in/raintank/schema.v0"
	"gopkg.in/raintank/schema.v0/msg"
)

var (
	Publisher            *Tsdb
	maxMetricsPerPayload = 50000
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
	Events       chan *schema.ProbeEvent
	triggerFlush chan struct{}
	LastFlush    time.Time
	closeChan    chan struct{}
	dataChan     chan tsdbData
}

func NewTsdb(u *url.URL, apiKey string, concurrency int) *Tsdb {
	t := &Tsdb{
		Metrics:      make([]*schema.MetricData, 0),
		triggerFlush: make(chan struct{}),
		Events:       make(chan *schema.ProbeEvent, 1000),
		Url:          u,
		ApiKey:       apiKey,
		Concurrency:  concurrency,
		dataChan:     make(chan tsdbData, 1000),
	}
	go t.Run()
	return t
}

func (t *Tsdb) Add(metrics []*schema.MetricData) {
	t.Lock()
	t.Metrics = append(t.Metrics, metrics...)
	if len(t.Metrics) > maxMetricsPerPayload {
		ticker := time.NewTicker(time.Second)
		pre := time.Now()
	FLUSHLOOP:
		for {
			select {
			case <-ticker.C:
				wait := time.Since(pre)
				log.Warn("unable to flush metrics fast enough. waited %f seconds", wait.Seconds())
			case t.triggerFlush <- struct{}{}:
				ticker.Stop()
				break FLUSHLOOP
			}
		}
	}
	t.Unlock()
}
func (t *Tsdb) AddEvent(event *schema.ProbeEvent) {
	t.Events <- event
}

func (t *Tsdb) Flush() {
	t.Lock()
	if len(t.Metrics) == 0 {
		t.Unlock()
		return
	}
	t.LastFlush = time.Now()
	metrics := make([]*schema.MetricData, len(t.Metrics))
	copy(metrics, t.Metrics)
	t.Metrics = t.Metrics[:0]
	t.Unlock()
	// Write the metrics to our HTTP server.
	log.Debug("writing metrics to API", "count", len(metrics))
	id := t.LastFlush.UnixNano()
	body, err := msg.CreateMsg(metrics, id, msg.FormatMetricDataArrayMsgp)
	if err != nil {
		log.Error(3, "unable to convert metrics to MetricDataArrayMsgp.", "error", err)
		return
	}
	t.dataChan <- tsdbData{Path: "metrics", Body: body}
}

func (t *Tsdb) Run() {
	for i := 0; i < t.Concurrency; i++ {
		go t.sendData()
	}
	ticker := time.NewTicker(time.Second)
	for {
		select {
		case <-ticker.C:
			t.Flush()
		case <-t.triggerFlush:
			t.Flush()
		case e := <-t.Events:
			t.SendEvent(e)
		case <-t.closeChan:
			close(t.dataChan)
			return
		}
	}
}
func (t *Tsdb) Close() {
	t.triggerFlush <- struct{}{}
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
			bytesSent += body.Len()
			req, err := http.NewRequest("POST", u, body)
			if err != nil {
				log.Error(3, "failed to create request payload. ", err)
				break
			}
			req.Header.Set("Content-Type", "rt-metric-binary-snappy")
			req.Header.Set("Authorization", "Bearer "+t.ApiKey)

			sent := false
			for !sent {
				if err := send(req); err != nil {
					log.Error(3, err.Error())
					time.Sleep(time.Second)
					body.Reset()
					snappyBody := snappy.NewWriter(body)
					snappyBody.Write(data.Body)
					snappyBody.Close()
				} else {
					sent = true
				}
			}
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

func (t *Tsdb) SendEvent(event *schema.ProbeEvent) {
	id := time.Now().UnixNano()
	body, err := msg.CreateProbeEventMsg(event, id, msg.FormatProbeEventMsgp)
	if err != nil {
		log.Error(3, "Unable to convert event to ProbeEventMsgp.", "error", err)
		return
	}
	t.dataChan <- tsdbData{Path: "events", Body: body}
}
