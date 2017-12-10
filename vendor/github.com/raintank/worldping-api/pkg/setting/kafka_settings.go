package setting

type KafkaSettings struct {
	Enabled bool
	Brokers string
	Topic   string
}

func readKafkaSettings() {
	sec := Cfg.Section("kafka")
	Kafka.Enabled = sec.Key("enabled").MustBool(false)
	Kafka.Brokers = sec.Key("brokers").MustString("localhost:9092")
	Kafka.Topic = sec.Key("topic").MustString("worldping")
}
