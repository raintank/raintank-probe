// Copyright 2014 Unknwon
// Copyright 2014 Torkel Ödegaard

package setting

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/ini.v1"

	"github.com/grafana/grafana/pkg/util"
	"github.com/raintank/worldping-api/pkg/log"
)

type Scheme string

const (
	HTTP  Scheme = "http"
	HTTPS Scheme = "https"
)

const (
	DEV  string = "development"
	PROD string = "production"
	TEST string = "test"
)

var (
	// App settings.
	Env        string = DEV
	InstanceId string
	AppUrl     string
	AppSubUrl  string

	// build
	BuildVersion string
	BuildCommit  string
	BuildStamp   int64

	// Paths
	LogsPath string
	HomePath string
	DataPath string

	// Log settings.
	LogModes   []string
	LogConfigs []util.DynMap

	// Http server options
	Protocol           Scheme
	HttpAddr, HttpPort string
	SshPort            int
	CertFile, KeyFile  string
	RouterLogging      bool
	StaticRootPath     string
	EnableGzip         bool

	// Http auth
	AdminKey string

	// Global setting objects.
	Cfg          *ini.File
	ConfRootPath string

	//Raintank Graphite Backend
	ElasticsearchUrl string
	TsdbUrl          string

	// for logging purposes
	configFiles                  []string
	appliedCommandLineProperties []string
	appliedEnvOverrides          []string

	StatsdEnabled   bool
	StatsdAddr      string
	StatsdType      string
	ProfileHeapMB   int
	ProfileHeapWait int
	ProfileHeapDir  string

	Kafka KafkaSettings

	Alerting AlertingSettings

	// SMTP email settings
	Smtp SmtpSettings

	// QUOTA
	Quota QuotaSettings
)

type CommandLineArgs struct {
	Config   string
	HomePath string
	Args     []string
}

func init() {
	log.NewLogger(0, "console", `{"level": 0, "formatting":true}`)
}

func parseAppUrlAndSubUrl(section *ini.Section) (string, string) {
	appUrl := section.Key("root_url").MustString("http://localhost:3000/")
	if appUrl[len(appUrl)-1] != '/' {
		appUrl += "/"
	}

	// Check if has app suburl.
	url, err := url.Parse(appUrl)
	if err != nil {
		log.Fatal(4, "Invalid root_url(%s): %s", appUrl, err)
	}
	appSubUrl := strings.TrimSuffix(url.Path, "/")

	return appUrl, appSubUrl
}

func ToAbsUrl(relativeUrl string) string {
	return AppUrl + relativeUrl
}

func applyEnvVariableOverrides() {
	appliedEnvOverrides = make([]string, 0)
	for _, section := range Cfg.Sections() {
		for _, key := range section.Keys() {
			sectionName := strings.ToUpper(strings.Replace(section.Name(), ".", "_", -1))
			keyName := strings.ToUpper(strings.Replace(key.Name(), ".", "_", -1))
			envKey := fmt.Sprintf("WP_%s_%s", sectionName, keyName)
			envValue := os.Getenv(envKey)

			if len(envValue) > 0 {
				key.SetValue(envValue)
				appliedEnvOverrides = append(appliedEnvOverrides, fmt.Sprintf("%s=%s", envKey, envValue))
			}
		}
	}
}

func applyCommandLineDefaultProperties(props map[string]string) {
	appliedCommandLineProperties = make([]string, 0)
	for _, section := range Cfg.Sections() {
		for _, key := range section.Keys() {
			keyString := fmt.Sprintf("default.%s.%s", section.Name(), key.Name())
			value, exists := props[keyString]
			if exists {
				key.SetValue(value)
				appliedCommandLineProperties = append(appliedCommandLineProperties, fmt.Sprintf("%s=%s", keyString, value))
			}
		}
	}
}

func applyCommandLineProperties(props map[string]string) {
	for _, section := range Cfg.Sections() {
		for _, key := range section.Keys() {
			keyString := fmt.Sprintf("%s.%s", section.Name(), key.Name())
			value, exists := props[keyString]
			if exists {
				key.SetValue(value)
				appliedCommandLineProperties = append(appliedCommandLineProperties, fmt.Sprintf("%s=%s", keyString, value))
			}
		}
	}
}

func getCommandLineProperties(args []string) map[string]string {
	props := make(map[string]string)

	for _, arg := range args {
		if !strings.HasPrefix(arg, "cfg:") {
			continue
		}

		trimmed := strings.TrimPrefix(arg, "cfg:")
		parts := strings.Split(trimmed, "=")
		if len(parts) != 2 {
			log.Fatal(3, "Invalid command line argument", arg)
			return nil
		}

		props[parts[0]] = parts[1]
	}
	return props
}

func makeAbsolute(path string, root string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(root, path)
}

func evalEnvVarExpression(value string) string {
	regex := regexp.MustCompile(`\${(\w+)}`)
	return regex.ReplaceAllStringFunc(value, func(envVar string) string {
		envVar = strings.TrimPrefix(envVar, "${")
		envVar = strings.TrimSuffix(envVar, "}")
		envValue := os.Getenv(envVar)
		return envValue
	})
}

func evalConfigValues() {
	for _, section := range Cfg.Sections() {
		for _, key := range section.Keys() {
			key.SetValue(evalEnvVarExpression(key.Value()))
		}
	}
}

func loadSpecifedConfigFile(configFile string) {
	if configFile == "" {
		configFile = filepath.Join(HomePath, "conf/custom.ini")
		// return without error if custom file does not exist
		if !pathExists(configFile) {
			return
		}
	}

	userConfig, err := ini.Load(configFile)
	userConfig.BlockMode = false
	if err != nil {
		log.Fatal(3, "Failed to parse %v, %v", configFile, err)
	}

	for _, section := range userConfig.Sections() {
		for _, key := range section.Keys() {
			if key.Value() == "" {
				continue
			}

			defaultSec, err := Cfg.GetSection(section.Name())
			if err != nil {
				log.Error(3, "Unknown config section %s defined in %s", section.Name(), configFile)
				continue
			}
			defaultKey, err := defaultSec.GetKey(key.Name())
			if err != nil {
				log.Error(3, "Unknown config key %s defined in section %s, in file %s", key.Name(), section.Name(), configFile)
				continue
			}
			defaultKey.SetValue(key.Value())
		}
	}

	configFiles = append(configFiles, configFile)
}

func loadConfiguration(args *CommandLineArgs) {
	var err error

	// load config defaults
	defaultConfigFile := path.Join(HomePath, "conf/defaults.ini")
	configFiles = append(configFiles, defaultConfigFile)

	Cfg, err = ini.Load(defaultConfigFile)

	if err != nil {
		log.Fatal(3, "Failed to parse defaults.ini, %v", err)
	}
	Cfg.BlockMode = false

	// command line props
	commandLineProps := getCommandLineProperties(args.Args)
	// load default overrides
	applyCommandLineDefaultProperties(commandLineProps)

	// init logging before specific config so we can log errors from here on
	DataPath = makeAbsolute(Cfg.Section("paths").Key("data").String(), HomePath)
	initLogging(args)

	// load specified config file
	loadSpecifedConfigFile(args.Config)

	// apply environment overrides
	applyEnvVariableOverrides()

	// apply command line overrides
	applyCommandLineProperties(commandLineProps)

	// evaluate config values containing environment variables
	evalConfigValues()

	// update data path and logging config
	DataPath = makeAbsolute(Cfg.Section("paths").Key("data").String(), HomePath)
	initLogging(args)
}

func pathExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return false
}

func setHomePath(args *CommandLineArgs) {
	if args.HomePath != "" {
		HomePath = args.HomePath
		return
	}

	HomePath, _ = filepath.Abs(".")
	// check if homepath is correct
	if pathExists(filepath.Join(HomePath, "conf/defaults.ini")) {
		return
	}

	// try down one path
	if pathExists(filepath.Join(HomePath, "../conf/defaults.ini")) {
		HomePath = filepath.Join(HomePath, "../")
	}
}

func NewConfigContext(args *CommandLineArgs) error {
	setHomePath(args)
	loadConfiguration(args)

	Env = Cfg.Section("").Key("app_mode").MustString("development")
	InstanceId = Cfg.Section("").Key("instance_id").MustString("default")

	server := Cfg.Section("server")
	AppUrl, AppSubUrl = parseAppUrlAndSubUrl(server)

	Protocol = HTTP
	if server.Key("protocol").MustString("http") == "https" {
		Protocol = HTTPS
		CertFile = server.Key("cert_file").String()
		KeyFile = server.Key("cert_key").String()
	}

	HttpAddr = server.Key("http_addr").MustString("0.0.0.0")
	HttpPort = server.Key("http_port").MustString("3000")
	RouterLogging = server.Key("router_logging").MustBool(false)
	EnableGzip = server.Key("enable_gzip").MustBool(false)
	StaticRootPath = makeAbsolute(server.Key("static_root_path").String(), HomePath)

	AdminKey = server.Key("admin_key").String()

	ElasticsearchUrl = Cfg.Section("raintank").Key("elasticsearch_url").MustString("http://localhost:9200/")
	if ElasticsearchUrl[len(ElasticsearchUrl)-1] != '/' {
		ElasticsearchUrl += "/"
	}
	_, err := url.Parse(ElasticsearchUrl)
	if err != nil {
		log.Fatal(4, "Invalid elasticsearch_url(%s): %s", ElasticsearchUrl, err)
	}

	TsdbUrl = Cfg.Section("raintank").Key("tsdb_url").MustString("http://tsdb-gw/")
	if TsdbUrl[len(TsdbUrl)-1] != '/' {
		TsdbUrl += "/"
	}
	_, err = url.Parse(TsdbUrl)
	if err != nil {
		log.Fatal(4, "Invalid tsdb_url(%s): %s", TsdbUrl, err)
	}

	telemetry := Cfg.Section("telemetry")
	StatsdEnabled = telemetry.Key("statsd_enabled").MustBool(false)
	StatsdAddr = telemetry.Key("statsd_addr").String()
	StatsdType = telemetry.Key("statsd_type").String()
	ProfileHeapMB = telemetry.Key("profile_heap_MB").MustInt(0)
	ProfileHeapWait = telemetry.Key("profile_heap_wait").MustInt(3600)
	ProfileHeapDir = telemetry.Key("profile_heap_dir").MustString("/tmp")

	readKafkaSettings()
	readAlertingSettings()
	readSmtpSettings()
	readQuotaSettings()
	return nil
}

var logLevels = map[string]int{
	"Trace":    0,
	"Debug":    1,
	"Info":     2,
	"Warn":     3,
	"Error":    4,
	"Critical": 5,
}

func initLogging(args *CommandLineArgs) {
	//close any existing log handlers.
	log.Close()
	// Get and check log mode.
	LogModes = strings.Split(Cfg.Section("log").Key("mode").MustString("console"), ",")
	LogsPath = makeAbsolute(Cfg.Section("paths").Key("logs").String(), HomePath)

	LogConfigs = make([]util.DynMap, len(LogModes))
	for i, mode := range LogModes {
		mode = strings.TrimSpace(mode)
		sec, err := Cfg.GetSection("log." + mode)
		if err != nil {
			log.Fatal(4, "Unknown log mode: %s", mode)
		}

		// Log level.
		levelName := Cfg.Section("log."+mode).Key("level").In("Trace",
			[]string{"Trace", "Debug", "Info", "Warn", "Error", "Critical"})
		level, ok := logLevels[levelName]
		if !ok {
			log.Fatal(4, "Unknown log level: %s", levelName)
		}

		// Generate log configuration.
		switch mode {
		case "console":
			formatting := sec.Key("formatting").MustBool(true)
			LogConfigs[i] = util.DynMap{
				"level":      level,
				"formatting": formatting,
			}
		case "file":
			logPath := sec.Key("file_name").MustString(filepath.Join(LogsPath, "worldping-api.log"))
			os.MkdirAll(filepath.Dir(logPath), os.ModePerm)
			LogConfigs[i] = util.DynMap{
				"level":    level,
				"filename": logPath,
				"rotate":   sec.Key("log_rotate").MustBool(true),
				"maxlines": sec.Key("max_lines").MustInt(1000000),
				"maxsize":  1 << uint(sec.Key("max_size_shift").MustInt(28)),
				"daily":    sec.Key("daily_rotate").MustBool(true),
				"maxdays":  sec.Key("max_days").MustInt(7),
			}
		case "conn":
			LogConfigs[i] = util.DynMap{
				"level":          level,
				"reconnectOnMsg": sec.Key("reconnect_on_msg").MustBool(),
				"reconnect":      sec.Key("reconnect").MustBool(),
				"net":            sec.Key("protocol").In("tcp", []string{"tcp", "unix", "udp"}),
				"addr":           sec.Key("addr").MustString(":7020"),
			}
		case "smtp":
			LogConfigs[i] = util.DynMap{
				"level":     level,
				"user":      sec.Key("user").MustString("example@example.com"),
				"passwd":    sec.Key("passwd").MustString("******"),
				"host":      sec.Key("host").MustString("127.0.0.1:25"),
				"receivers": sec.Key("receivers").MustString("[]"),
				"subject":   sec.Key("subject").MustString("Diagnostic message from serve"),
			}
		case "database":
			LogConfigs[i] = util.DynMap{
				"level":  level,
				"driver": sec.Key("driver").String(),
				"conn":   sec.Key("conn").String(),
			}
		}

		cfgJsonBytes, _ := json.Marshal(LogConfigs[i])
		log.NewLogger(Cfg.Section("log").Key("buffer_len").MustInt64(10000), mode, string(cfgJsonBytes))
	}
}

func LogConfigurationInfo() {
	var text bytes.Buffer
	text.WriteString("Configuration Info\n")

	text.WriteString("Config files:\n")
	for i, file := range configFiles {
		text.WriteString(fmt.Sprintf("  [%d]: %s\n", i, file))
	}

	if len(appliedCommandLineProperties) > 0 {
		text.WriteString("Command lines overrides:\n")
		for i, prop := range appliedCommandLineProperties {
			text.WriteString(fmt.Sprintf("  [%d]: %s\n", i, prop))
		}
	}

	if len(appliedEnvOverrides) > 0 {
		text.WriteString("\tEnvironment variables used:\n")
		for i, prop := range appliedEnvOverrides {
			text.WriteString(fmt.Sprintf("  [%d]: %s\n", i, prop))
		}
	}

	text.WriteString("Paths:\n")
	text.WriteString(fmt.Sprintf("  home: %s\n", HomePath))
	text.WriteString(fmt.Sprintf("  data: %s\n", DataPath))
	text.WriteString(fmt.Sprintf("  logs: %s\n", LogsPath))

	log.Info(text.String())
}
