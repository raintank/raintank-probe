package setting

import (
	"net/url"

	"github.com/raintank/worldping-api/pkg/log"
)

type AlertingSettings struct {
	Enabled              bool
	Topic                string
	Distributed          bool
	TickQueueSize        int
	InternalJobQueueSize int
	ExecutorLRUSize      int
	EnableScheduler      bool
	EnableWorker         bool
	Executors            int
	GraphiteUrl          string
}

func readAlertingSettings() {
	alerting := Cfg.Section("alerting")
	Alerting.Enabled = alerting.Key("enabled").MustBool(false)
	Alerting.Distributed = alerting.Key("distributed").MustBool(false)
	Alerting.Topic = alerting.Key("topic").MustString("worldping-alerts")
	Alerting.TickQueueSize = alerting.Key("tickqueue_size").MustInt(0)
	Alerting.InternalJobQueueSize = alerting.Key("internal_jobqueue_size").MustInt(0)

	Alerting.ExecutorLRUSize = alerting.Key("executor_lru_size").MustInt(0)
	Alerting.EnableScheduler = alerting.Key("enable_scheduler").MustBool(true)
	Alerting.EnableWorker = alerting.Key("enable_worker").MustBool(true)

	Alerting.GraphiteUrl = alerting.Key("graphite_url").MustString("http://localhost:8888/")
	if Alerting.GraphiteUrl[len(Alerting.GraphiteUrl)-1] != '/' {
		Alerting.GraphiteUrl += "/"
	}
	// Check if has app suburl.
	_, err := url.Parse(Alerting.GraphiteUrl)
	if err != nil {
		log.Fatal(4, "Invalid graphite_url(%s): %s", Alerting.GraphiteUrl, err)
	}

	if Alerting.Distributed && !Kafka.Enabled {
		log.Fatal(4, "Kafka must be enabled to use distributed alerting.")

	}
}
