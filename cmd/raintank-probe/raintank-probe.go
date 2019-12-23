package main

import (
	"flag"
	"fmt"
	"net"
	"net/http"
	_ "net/http/pprof"
	"net/url"
	"os"
	"os/signal"
	"path"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
	"github.com/grafana/metrictank/stats"
	gosocketio "github.com/gsocket-io/golang-socketio"
	"github.com/gsocket-io/golang-socketio/transport"
	"github.com/raintank/metrictank/logger"
	"github.com/raintank/raintank-probe/checks"
	"github.com/raintank/raintank-probe/probe"
	"github.com/raintank/raintank-probe/publisher"
	"github.com/raintank/raintank-probe/scheduler"
	m "github.com/raintank/worldping-api/pkg/models"
	"github.com/rakyll/globalconf"
	log "github.com/sirupsen/logrus"
)

const Version int = 1

var (
	GitHash     = "(none)"
	showVersion = flag.Bool("version", false, "print version string")
	logLevel    = flag.Int("log-level", 2, "log level. 0=TRACE|1=DEBUG|2=INFO|3=WARN|4=ERROR|5=FATAL|6=PANIC")
	confFile    = flag.String("config", "/etc/raintank/probe.ini", "configuration file path")

	serverAddr  = flag.String("server-url", "ws://localhost:80/", "address of worldping-api server")
	tsdbAddr    = flag.String("tsdb-url", "http://localhost:80/", "address of tsdb server")
	nodeName    = flag.String("name", "", "agent-name")
	apiKey      = flag.String("api-key", "not_very_secret_key", "Api Key")
	concurrency = flag.Int("concurrency", 5, "concurrency number of requests to TSDB.")
	healthHosts = flag.String("health-hosts", "google.com,youtube.com,facebook.com,twitter.com,wikipedia.com", "comma separted list of hosts to ping to determin network health of this probe.")

	statsEnabled    = flag.Bool("stats-enabled", false, "enable sending graphite messages for instrumentation")
	statsPrefix     = flag.String("stats-prefix", "raintank-probe.stats.$hostname", "stats prefix (will add trailing dot automatically if needed)")
	statsAddr       = flag.String("stats-addr", "localhost:2003", "graphite address")
	statsInterval   = flag.Int("stats-interval", 10, "interval in seconds to send statistics")
	statsBufferSize = flag.Int("stats-buffer-size", 20000, "how many messages (holding all measurements from one interval) to buffer up in case graphite endpoint is unavailable.")
	statsTimeout    = flag.Duration("stats-timeout", time.Second*10, "timeout after which a write is considered not successful")

	// healthz endpoint
	healthzListenAddr = flag.String("healthz-listen-addr", "localhost:7180", "address to listen on for healthz http api.")

	MonitorTypes map[string]m.MonitorTypeDTO

	wsTransport = transport.GetDefaultWebsocketTransport()

	// metrics
	controllerConnected = stats.NewGauge32("controller.connected")
	eventsReceived      = stats.NewCounterRate32("handlers.events.received")
)

func main() {
	flag.Parse()
	// Set 'cfile' here if *confFile exists, because we should only try and
	// parse the conf file if it exists. If we try and parse the default
	// conf file location when it's not there, we (unsurprisingly) get a
	// panic.
	var cfile string
	if _, err := os.Stat(*confFile); err == nil {
		cfile = *confFile
	}

	// Still parse globalconf, though, even if the config file doesn't exist
	// because we want to be able to use environment variables.
	conf, err := globalconf.NewWithOptions(&globalconf.Options{
		Filename:  cfile,
		EnvPrefix: "RTPROBE_",
	})
	if err != nil {
		panic(fmt.Sprintf("error with configuration file: %s", err))
	}
	conf.ParseAll()

	initLogger(*logLevel)

	if *showVersion {
		fmt.Printf("raintank-probe (built with %s, git hash %s)\n", runtime.Version(), GitHash)
		return
	}

	if *statsEnabled {
		stats.NewMemoryReporter()
		hostname, _ := os.Hostname()
		prefix := strings.Replace(*statsPrefix, "$hostname", strings.Replace(hostname, ".", "_", -1), -1)
		stats.NewGraphite(prefix, *statsAddr, *statsInterval, *statsBufferSize, *statsTimeout)
	} else {
		stats.NewDevnull()
	}

	if *nodeName == "" {
		log.Fatal("name must be set.")
	}

	checks.InitPinger()

	jobScheduler := scheduler.New(*healthHosts)
	go jobScheduler.CheckHealth()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	controllerUrl, err := url.Parse(*serverAddr)
	if err != nil {
		log.Fatalf("unable to parse server-url: %s", err)
	}
	controllerUrl.Path = path.Clean(controllerUrl.Path + "/socket.io")
	version := strings.Split(GitHash, "-")[0]
	controllerUrl.RawQuery = fmt.Sprintf("EIO=3&transport=websocket&apiKey=%s&name=%s&version=%s", *apiKey, url.QueryEscape(*nodeName), version)

	if controllerUrl.Scheme != "ws" && controllerUrl.Scheme != "wss" {
		log.Fatalf("invalid server-url.  scheme must be ws or wss. was %s", controllerUrl.Scheme)
	}

	tsdbUrl, err := url.Parse(*tsdbAddr)
	if err != nil {
		log.Fatalf("unable to parse tsdb-url: %s", err)
	}
	if !strings.HasPrefix(tsdbUrl.Path, "/") {
		tsdbUrl.Path += "/"
	}
	publisher.Init(tsdbUrl, *apiKey, *concurrency)

	wsTransport.Dialer = &websocket.Dialer{
		Proxy:            http.ProxyFromEnvironment,
		HandshakeTimeout: time.Minute,
	}
	go connectController(controllerUrl, jobScheduler, interrupt, true)

	healthz := NewHealthz(jobScheduler)
	go healthz.Run()
	//wait for interupt Signal.
	<-interrupt
	log.Info("interrupt")
	healthz.Stop()
	jobScheduler.Close()
	publisher.Stop()
	checks.GlobalPinger.Stop()
	log.Info("exiting")
	return
}

func connectController(controllerUrl *url.URL, jobScheduler *scheduler.Scheduler, interrupt chan os.Signal, failOncConnect bool) {
	log.Infof("attempting to connect to controller at %s", controllerUrl.Hostname())
	connected := false
	controllerConnected.Set(0)
	var client *gosocketio.Client
	var err error
	eventChan := make(chan string, 2)
	for !connected {
		client, err = gosocketio.Dial(controllerUrl.String(), wsTransport)
		if err != nil {
			log.Error(3, err.Error())
			if failOncConnect {
				log.Errorf("unable to connect to controller on url %s: %s", controllerUrl.String(), err)
				close(interrupt)
			} else {
				log.Error("failed to connect to controller. will retry.")
				time.Sleep(time.Second * 2)
			}
		} else {
			connected = true
			controllerConnected.Set(1)
			log.Info("Connected to controller")
			bindHandlers(client, controllerUrl, jobScheduler, interrupt, eventChan)
		}
	}

	maxInactivity := time.Minute * 30
	timer := time.NewTimer(maxInactivity)
	for {
		select {
		case <-interrupt:
			client.Close()
			return
		case <-timer.C:
			// we have not received on the notifyRefresh channel.
			// if the client is alive, close it. Otherwise reconnect
			if client.IsAlive() {
				log.Warn("no refresh received for maxInactivity time. disconnecting from controller.")
				client.Close()
				// once closed a "disconnected" event will be emitted.
				// we will then re-establish the connection.
			} else {
				go connectController(controllerUrl, jobScheduler, interrupt, false)
				return
			}
		case event := <-eventChan:
			switch event {
			case "disconnected":
				go connectController(controllerUrl, jobScheduler, interrupt, false)
				return
			case "refresh":
				log.Debug("refresh event received on eventChan")
			}
		}
		if !timer.Stop() {
			<-timer.C
		}
		timer.Reset(maxInactivity)
	}
}

func bindHandlers(client *gosocketio.Client, controllerUrl *url.URL, jobScheduler *scheduler.Scheduler, interrupt chan os.Signal, eventChan chan string) {
	if !client.IsAlive() {
		log.Error("Connection to controller closed before binding handlers.")
		eventChan <- "disconnected"
		return
	}
	client.On(gosocketio.OnDisconnection, func(c *gosocketio.Channel) {
		eventsReceived.Inc()
		log.Error("Disconnected from controller.")
		eventChan <- "disconnected"
	})
	client.On("refresh", func(c *gosocketio.Channel, checks []*m.CheckWithSlug) {
		eventsReceived.Inc()
		eventChan <- "refresh"
		jobScheduler.Refresh(checks)
	})
	client.On("created", func(c *gosocketio.Channel, check m.CheckWithSlug) {
		eventsReceived.Inc()
		jobScheduler.Create(&check)
	})
	client.On("updated", func(c *gosocketio.Channel, check m.CheckWithSlug) {
		eventsReceived.Inc()
		jobScheduler.Update(&check)
	})
	client.On("removed", func(c *gosocketio.Channel, check m.CheckWithSlug) {
		eventsReceived.Inc()
		jobScheduler.Remove(&check)
	})

	client.On("ready", func(c *gosocketio.Channel, event m.ProbeReadyPayload) {
		eventsReceived.Inc()
		log.Infof("server sent ready event. ProbeId=%d", event.Collector.Id)
		probe.Self = event.Collector

		queryParams := controllerUrl.Query()
		queryParams["lastSocketId"] = []string{event.SocketId}
		controllerUrl.RawQuery = queryParams.Encode()

	})
	client.On("error", func(c *gosocketio.Channel, reason string) {
		eventsReceived.Inc()
		log.Errorf("Controller emitted an error. %s", reason)
		close(interrupt)
	})
}

type Healthz struct {
	listener     net.Listener
	jobScheduler *scheduler.Scheduler
}

// runs a HTTP server, accepting requests to /ready and /alive which reports the
// readiness/liveness of the probe
func NewHealthz(jobScheduler *scheduler.Scheduler) *Healthz {
	// define our own listner so we can call Close on it
	l, err := net.Listen("tcp", *healthzListenAddr)
	if err != nil {
		log.Fatal(4, err.Error())
	}
	return &Healthz{
		listener:     l,
		jobScheduler: jobScheduler,
	}
}

func (h *Healthz) Run() {
	http.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		healthy := h.jobScheduler.IsHealthy()
		if healthy {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte("Not Ready"))
		}
	})

	http.HandleFunc("/alive", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	srv := http.Server{
		Addr: *healthzListenAddr,
	}
	err := srv.Serve(h.listener)
	if err != nil {
		log.Info(err.Error())
	}
}

func (h *Healthz) Stop() {
	h.listener.Close()
	log.Info("healthz listener closed")
}

func initLogger(level int) {
	formatter := &logger.TextFormatter{}
	formatter.TimestampFormat = "2006-01-02 15:04:05.000"
	log.SetFormatter(formatter)

	// map old loglevel settings to logrus settings.
	// 0=TRACE|1=DEBUG|2=INFO|3=WARN|4=ERROR|5=FATAL|6=PANIC"
	switch level {
	case 0:
		log.SetLevel(log.TraceLevel)
	case 1:
		log.SetLevel(log.DebugLevel)
	case 2:
		log.SetLevel(log.InfoLevel)
	case 3:
		log.SetLevel(log.WarnLevel)
	case 4:
		log.SetLevel(log.ErrorLevel)
	case 5:
		log.SetLevel(log.FatalLevel)
	case 6:
		log.SetLevel(log.PanicLevel)
	default:
		log.Fatalf("unknown log level %d", level)
	}

	log.Infof("logging level set to '%s'", log.GetLevel().String())
}
