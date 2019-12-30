package main

import (
	"flag"
	"fmt"
	_ "net/http/pprof"
	"net/url"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/grafana/metrictank/stats"
	"github.com/raintank/metrictank/logger"
	"github.com/raintank/raintank-probe/checks"
	"github.com/raintank/raintank-probe/controller"
	"github.com/raintank/raintank-probe/healthz"
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

	tsdbUrl, err := url.Parse(*tsdbAddr)
	if err != nil {
		log.Fatalf("unable to parse tsdb-url: %s", err)
	}
	if !strings.HasPrefix(tsdbUrl.Path, "/") {
		tsdbUrl.Path += "/"
	}
	publisher.Init(tsdbUrl, *apiKey, *concurrency)

	// init the GlobalPinger. go-pinger uses raw sockets, so if the process does not have CAP_NET
	// privileges, the process will panic.
	checks.InitPinger()

	jobScheduler := scheduler.New(*healthHosts)
	go jobScheduler.CheckHealth()

	healthz := healthz.NewHealthz(jobScheduler, *healthzListenAddr)

	version := strings.Split(GitHash, "-")[0]
	controllerCfg := &controller.ControllerConfig{
		ServerAddr:   *serverAddr,
		ApiKey:       *apiKey,
		NodeName:     *nodeName,
		Version:      version,
		JobScheduler: jobScheduler,
	}
	controller := controller.NewController(controllerCfg)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	//wait for interupt Signal.
	<-interrupt
	log.Info("Shutting down")
	controller.Stop()
	healthz.Stop()
	jobScheduler.Close()
	publisher.Stop()
	checks.GlobalPinger.Stop()
	log.Info("exiting")
	return
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
