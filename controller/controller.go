package controller

import (
	"fmt"
	"net/http"
	"net/url"
	"path"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/grafana/metrictank/stats"
	gosocketio "github.com/gsocket-io/golang-socketio"
	"github.com/gsocket-io/golang-socketio/transport"
	"github.com/jpillora/backoff"
	"github.com/raintank/raintank-probe/probe"
	"github.com/raintank/raintank-probe/scheduler"
	m "github.com/raintank/worldping-api/pkg/models"
	log "github.com/sirupsen/logrus"
)

var (
	// metrics
	controllerConnected = stats.NewGauge32("controller.connected")
	eventsReceived      = stats.NewCounterRate32("handlers.events.received")
)

type ControllerConfig struct {
	ServerAddr   string
	ApiKey       string
	NodeName     string
	Version      string
	JobScheduler *scheduler.Scheduler
}

type Controller struct {
	sync.Mutex

	URL          *url.URL
	readyChan    chan m.ProbeReadyPayload
	shutdown     chan struct{}
	jobScheduler *scheduler.Scheduler
}

func NewController(cfg *ControllerConfig) *Controller {
	controllerURL, err := url.Parse(cfg.ServerAddr)
	if err != nil {
		log.Fatalf("unable to parse server-url: %s", err)
	}
	if controllerURL.Scheme != "ws" && controllerURL.Scheme != "wss" {
		log.Fatalf("invalid server-url.  scheme must be ws or wss. was %s", controllerURL.Scheme)
	}

	controllerURL.Path = path.Clean(controllerURL.Path + "/socket.io")
	controllerURL.RawQuery = fmt.Sprintf("EIO=3&transport=websocket&apiKey=%s&name=%s&version=%s", url.QueryEscape(cfg.ApiKey), url.QueryEscape(cfg.NodeName), url.QueryEscape(cfg.Version))

	c := &Controller{
		URL:          controllerURL,
		jobScheduler: cfg.JobScheduler,
		shutdown:     make(chan struct{}),
		readyChan:    make(chan m.ProbeReadyPayload),
	}
	go c.loop(true)
	return c
}

func (c *Controller) Stop() {
	close(c.shutdown)
}

// Address returns the full URL of the controller server, minus the apiKey
func (c *Controller) Address() string {
	u, _ := url.Parse(c.URL.String())
	queryParams := u.Query()
	queryParams.Del("apiKey")
	u.RawQuery = queryParams.Encode()
	return u.String()
}

// core control loop.
func (c *Controller) loop(panicOnConnectionFailure bool) {
	// use a mutex to ensure only one instance of the loop is running
	c.Lock()
	defer c.Unlock()
	log.Infof("attempting to connect to controller at %s", c.Address())
	wsTransport := transport.GetDefaultWebsocketTransport()
	wsTransport.Dialer = &websocket.Dialer{
		Proxy:            http.ProxyFromEnvironment,
		HandshakeTimeout: time.Minute,
	}
	connected := false
	controllerConnected.Set(0)
	var client *gosocketio.Client
	var err error

	b := &backoff.Backoff{
		Min:    time.Second,
		Max:    time.Minute,
		Factor: 2,
		Jitter: true,
	}
	for !connected {
		client, err = gosocketio.Dial(c.URL.String(), wsTransport)
		if err != nil {
			log.Errorf("failed to connect to controller. %s", err)
			if panicOnConnectionFailure {
				log.Fatalf("unable to connect to controller on url %s: %s", c.Address(), err)
			} else {
				log.Errorf("failed to connect to controller. %s", err)
				time.Sleep(b.Duration())
			}
		} else {
			connected = true
			controllerConnected.Set(1)
			log.Info("Connected to controller")
		}
	}
	// create a dedicated channel for this instance of the control loop.
	// when this function ends, we drain and close the channel
	eventChan := make(chan string, 2)
	defer drainEventChan(eventChan)
	c.bindHandlers(client, eventChan)

	maxInactivity := time.Minute * 30
	timer := time.NewTimer(maxInactivity)
	for {
		select {
		case <-c.shutdown:
			log.Info("controller loop exiting")
			client.Close()
			return
		case <-timer.C:
			log.Warn("no refresh received for maxInactivity time")
			// we have not received on the notifyRefresh channel.
			// if the client is alive, close it before trying to reconnect
			if client.IsAlive() {
				log.Info("closing connection to controller")
				client.Close()
				// once closed, a "disconnected" event will be emitted,
				// but we can just ignore it.
			}
			go c.loop(false)
			return
		case event := <-eventChan:
			switch event {
			case "disconnected":
				go c.loop(false)
				return
			case "refresh":
				log.Debug("refresh event received on eventChan")
			}
		case event := <-c.readyChan:
			log.Infof("server sent ready event. ProbeId=%d", event.Collector.Id)
			probe.Self = event.Collector

			// update c.URL so next time we connect we pass our current socketID as
			// lastSocketId query param
			queryParams := c.URL.Query()
			queryParams["lastSocketId"] = []string{event.SocketId}
			c.URL.RawQuery = queryParams.Encode()
		}
		if !timer.Stop() {
			<-timer.C
		}
		timer.Reset(maxInactivity)
	}
}

func (c *Controller) bindHandlers(client *gosocketio.Client, eventChan chan string) {
	if !client.IsAlive() {
		log.Error("Connection to controller closed before binding handlers.")
		eventChan <- "disconnected"
		return
	}
	client.On(gosocketio.OnDisconnection, func(gsc *gosocketio.Channel) {
		eventsReceived.Inc()
		log.Error("Disconnected from controller.")
		eventChan <- "disconnected"
	})
	client.On("refresh", func(gsc *gosocketio.Channel, checks []*m.CheckWithSlug) {
		eventsReceived.Inc()
		eventChan <- "refresh"
		c.jobScheduler.Refresh(checks)
	})
	client.On("created", func(gsc *gosocketio.Channel, check m.CheckWithSlug) {
		eventsReceived.Inc()
		c.jobScheduler.Create(&check)
	})
	client.On("updated", func(gsc *gosocketio.Channel, check m.CheckWithSlug) {
		eventsReceived.Inc()
		c.jobScheduler.Update(&check)
	})
	client.On("removed", func(gsc *gosocketio.Channel, check m.CheckWithSlug) {
		eventsReceived.Inc()
		c.jobScheduler.Remove(&check)
	})
	client.On("ready", func(gsc *gosocketio.Channel, event m.ProbeReadyPayload) {
		eventsReceived.Inc()
		c.readyChan <- event
	})
	client.On("error", func(gsc *gosocketio.Channel, reason string) {
		eventsReceived.Inc()
		log.Fatalf("Controller emitted an error. %s", reason)
	})
}

func drainEventChan(eventChan chan string) {
	go func() {
		for {
			select {
			case <-time.After(time.Minute * 10):
				close(eventChan)
			case event, ok := <-eventChan:
				if !ok {
					return
				}
				log.Debugf("delayed %s event received", event)
			}
		}
	}()
}
