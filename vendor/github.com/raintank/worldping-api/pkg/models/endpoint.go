package models

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Typed errors
var (
	ErrEndpointNotFound = NewNotFoundError("Endpoint not found")
)

type Endpoint struct {
	Id      int64
	OrgId   int64
	Name    string
	Slug    string
	Created time.Time
	Updated time.Time
}

func (endpoint *Endpoint) UpdateSlug() {
	name := strings.ToLower(endpoint.Name)
	re := regexp.MustCompile("[^\\w ]+")
	re2 := regexp.MustCompile("\\s")
	endpoint.Slug = re2.ReplaceAllString(re.ReplaceAllString(name, "_"), "-")
}

type EndpointTag struct {
	Id         int64
	OrgId      int64
	EndpointId int64
	Tag        string
	Created    time.Time
}

type EndpointDTO struct {
	Id      int64     `json:"id"`
	OrgId   int64     `json:"orgId"`
	Name    string    `json:"name" binding:"Required"`
	Slug    string    `json:"slug"`
	Checks  []Check   `json:"checks"`
	Tags    []string  `json:"tags"`
	Created time.Time `json:"created"`
	Updated time.Time `json:"updated"`
}

type CheckType string

const (
	HTTP_CHECK  CheckType = "http"
	HTTPS_CHECK CheckType = "https"
	DNS_CHECK   CheckType = "dns"
	PING_CHECK  CheckType = "ping"
)

type Check struct {
	Id             int64                  `json:"id"`
	OrgId          int64                  `json:"orgId"`
	EndpointId     int64                  `json:"endpointId"`
	Route          *CheckRoute            `xorm:"JSON" json:"route"`
	Type           CheckType              `json:"type" binding:"Required,In(http,https,dns,ping)"`
	Frequency      int64                  `json:"frequency" binding:"Required,Range(10,300)"`
	Offset         int64                  `json:"offset"`
	Enabled        bool                   `json:"enabled"`
	State          CheckEvalResult        `json:"state"`
	StateChange    time.Time              `json:"stateChange"`
	StateCheck     time.Time              `json:"stateCheck"`
	Settings       map[string]interface{} `json:"settings" binding:"Required"`
	HealthSettings *CheckHealthSettings   `xorm:"JSON" json:"healthSettings"`
	Created        time.Time              `json:"created"`
	Updated        time.Time              `json:"updated"`
}

type CheckWithSlug struct {
	Check `xorm:"extends"`
	Slug  string `json:"endpointSlug"`
}

func (c Check) Validate(quotas []OrgQuotaDTO) error {
	// check route config
	if err := c.Route.Validate(); err != nil {
		return err
	}

	//check frequency
	validFreq := map[int64]bool{
		10:  true,
		30:  true,
		60:  true,
		120: true,
		300: true,
		600: true,
	}
	if _, ok := validFreq[c.Frequency]; !ok {
		return NewValidationError("Invalid frequency specified.")
	}

	//validate Settings.
	switch c.Type {
	case HTTP_CHECK:
		if err := c.validateHTTPSettings(quotas); err != nil {
			return err
		}
	case HTTPS_CHECK:
		if err := c.validateHTTPSSettings(quotas); err != nil {
			return err
		}
	case PING_CHECK:
		if err := c.validatePINGSettings(); err != nil {
			return err
		}
	case DNS_CHECK:
		if err := c.validateDNSSettings(); err != nil {
			return err
		}
	default:
		return NewValidationError(fmt.Sprintf("unknown check type. %s", c.Type))
	}
	return nil
}

type CheckHealthSettings struct {
	NumProbes     int                      `json:"num_collectors" binding:"Required"`
	Steps         int                      `json:"steps" binding:"Required"`
	Notifications CheckNotificationSetting `json:"notifications"`
}

type CheckNotificationSetting struct {
	Enabled   bool   `json:"enabled"`
	Addresses string `json:"addresses"`
}

type RouteType string

const (
	RouteByTags RouteType = "byTags"
	RouteByIds  RouteType = "byIds"
)

type RouteByIdIndex struct {
	Id      int64
	ProbeId int64
	CheckId int64
	Created time.Time
}
type RouteByTagIndex struct {
	Id      int64
	OrgId   int64
	CheckId int64
	Tag     string
	Created time.Time
}

var (
	InvalidRouteConfig = NewValidationError("Invalid route config")
	UnknownRouteType   = NewValidationError("unknown route type")
)

type CheckRoute struct {
	Type   RouteType              `json:"type" binding:"Required"`
	Config map[string]interface{} `json:"config"`
}

func (t *CheckRoute) UnmarshalJSON(body []byte) error {
	// use anonamous struct for intermediate decoding.
	// we need to decode the type to know how to decode
	// the config.
	firstPass := struct {
		Type   RouteType       `json:"type"`
		Config json.RawMessage `json:"config"`
	}{}

	err := json.Unmarshal(body, &firstPass)

	if err != nil {
		return err
	}
	config := make(map[string]interface{})

	t.Type = firstPass.Type
	switch firstPass.Type {
	case RouteByTags:
		c := make(map[string][]string)
		err = json.Unmarshal(firstPass.Config, &c)
		if err != nil {
			return err
		}
		for k, v := range c {
			config[k] = v
		}
	case RouteByIds:
		c := make(map[string][]int64)
		err = json.Unmarshal(firstPass.Config, &c)
		if err != nil {
			return err
		}
		for k, v := range c {
			config[k] = v
		}
	default:
		return UnknownRouteType
	}

	t.Config = config
	return err
}

func (r *CheckRoute) Validate() error {
	switch r.Type {
	case RouteByTags:
		if len(r.Config) != 1 {
			return InvalidRouteConfig
		}
		if _, ok := r.Config["tags"]; !ok {
			return InvalidRouteConfig
		}
	case RouteByIds:
		if len(r.Config) != 1 {
			return InvalidRouteConfig
		}
		if _, ok := r.Config["ids"]; !ok {
			return InvalidRouteConfig
		}
	default:
		return UnknownRouteType
	}
	return nil
}

// ----------------------
// COMMANDS
type DiscoverEndpointCmd struct {
	Name string `form:"name"`
}

// ---------------------
// QUERIES
type GetEndpointsQuery struct {
	OrgId   int64  `form:"-"`
	Name    string `form:"name"`
	Tag     string `form:"tag"`
	OrderBy string `form:"orderBy" binding:"In(name,slug,created,updated,)"`
}

//Alerting

type CheckState struct {
	Id      int64
	State   CheckEvalResult
	Updated time.Time // this protects against jobs running out of order.
	Checked time.Time
}

type CheckForAlertDTO struct {
	Id             int64
	OrgId          int64
	EndpointId     int64
	Slug           string
	Name           string
	Type           string
	Offset         int64
	Frequency      int64
	Enabled        bool
	State          CheckEvalResult
	StateChange    time.Time
	StateCheck     time.Time
	Settings       map[string]interface{} `xorm:"JSON"`
	HealthSettings *CheckHealthSettings   `xorm:"JSON"`
	Created        time.Time
	Updated        time.Time
}

func formatSize(size int64) string {
	if size > 1024*1024 {
		return fmt.Sprintf("%.2f MB", float64(size)/1024/1024)
	}
	if size > 1024 {
		return fmt.Sprintf("%.2f KB", float64(size)/1024)
	}
	return fmt.Sprintf("%d", size)
}

func (c Check) validateHTTPSettings(quotas []OrgQuotaDTO) error {
	settings := c.Settings

	requiredFields := map[string]string{
		"host": "string",
		"path": "string",
	}
	optFields := map[string]string{
		"port":          "number",
		"method":        "string",
		"headers":       "string",
		"expectRegex":   "string",
		"body":          "string",
		"timeout":       "number",
		"downloadLimit": "size",
		"ipversion":     "ipversion",
	}
	for field, dataType := range requiredFields {
		rawVal, ok := settings[field]
		if !ok {
			return NewValidationError(fmt.Sprintf("%s field missing from HTTP check", field))
		}
		switch dataType {
		case "string":
			value, ok := rawVal.(string)
			if !ok {
				return NewValidationError(fmt.Sprintf("%s field is invalid type. Expected string", field))
			}
			if value == "" {
				return NewValidationError(fmt.Sprintf("%s field missing from HTTP check", field))
			}
		}
	}

	for field, dataType := range optFields {
		rawVal, ok := settings[field]
		if !ok {
			continue
		}
		switch dataType {
		case "string":
			_, ok := rawVal.(string)
			if !ok {
				return NewValidationError(fmt.Sprintf("%s field is invalid type. Expected string", field))
			}
		case "number":
			value, ok := rawVal.(float64)
			if !ok {
				return NewValidationError(fmt.Sprintf("%s field is invalid type. Expected number", field))
			}
			if field == "timeout" {
				if value <= 0.0 || value > 10.0 {
					return NewValidationError(fmt.Sprintf("%s field is invalid. must be between 1 and 10", field))
				}
			}
			if field == "port" {
				settings[field] = int(value)
				if value < 1 || value > 65535 {
					return NewValidationError(fmt.Sprintf("%s field is invalid. must be between 1 and 65535", field))
				}
			}
		case "size":
			value := int64(0)

			switch rawVal.(type) {
			case float64:
				value = int64(rawVal.(float64))
			case int64:
				value = rawVal.(int64)
			case string:
				if rawVal.(string) == "" {
					delete(c.Settings, field)
					continue
				}

				re, err := regexp.Compile(`^(?i:(\d+)([km]?)b?)$`)
				if err != nil {
					return NewValidationError(fmt.Sprintf("%s field is invalid. error compiling size regexp", field))
				}

				matched := re.FindStringSubmatch(rawVal.(string))
				if matched == nil {
					return NewValidationError(fmt.Sprintf("%s field is invalid. must be number or size string", field))
				}

				value, err = strconv.ParseInt(matched[1], 10, 64)
				if strings.ToLower(matched[2]) == "m" {
					value = value * 1024 * 1024
				} else if strings.ToLower(matched[2]) == "k" {
					value = value * 1024
				}
			default:
				return NewValidationError(fmt.Sprintf("%s field is invalid. must be number or size string", field))
			}

			if field == "downloadLimit" {
				for _, quota := range quotas {
					if quota.Target == "downloadLimit" && value > int64(quota.Limit) {
						return NewValidationError(fmt.Sprintf("%s field is invalid. %s is over limit of %s", field, formatSize(value), formatSize(int64(quota.Limit))))
					}
				}
			}
		case "ipversion":
			version, ok := rawVal.(string)
			if !ok {
				return NewValidationError(fmt.Sprintf("%s field is invalid type. Expected ip version", field))
			}
			if !(version == "v4" || version == "v6" || version == "any") {
				return NewValidationError(fmt.Sprintf("%s field is invalid. Expected v4, v6 or any", field))
			}
		}
	}

	return nil

}
func (c Check) validateHTTPSSettings(quotas []OrgQuotaDTO) error {
	settings := c.Settings

	requiredFields := map[string]string{
		"host": "string",
		"path": "string",
	}
	optFields := map[string]string{
		"port":          "number",
		"method":        "string",
		"headers":       "string",
		"expectRegex":   "string",
		"validateCert":  "bool",
		"body":          "string",
		"timeout":       "number",
		"downloadLimit": "size",
		"ipversion":     "ipversion",
	}
	for field, dataType := range requiredFields {
		rawVal, ok := settings[field]
		if !ok {
			return NewValidationError(fmt.Sprintf("%s field missing from HTTPS check", field))
		}
		switch dataType {
		case "string":
			value, ok := rawVal.(string)
			if !ok {
				return NewValidationError(fmt.Sprintf("%s field is invalid type. Expected string", field))
			}
			if value == "" {
				return NewValidationError(fmt.Sprintf("%s field missing from HTTPS check", field))
			}
		}
	}

	for field, dataType := range optFields {
		rawVal, ok := settings[field]
		if !ok {
			continue
		}
		switch dataType {
		case "string":
			_, ok := rawVal.(string)
			if !ok {
				return NewValidationError(fmt.Sprintf("%s field is invalid type. Expected string", field))
			}
		case "number":
			value, ok := rawVal.(float64)
			if !ok {
				return NewValidationError(fmt.Sprintf("%s field is invalid type. Expected number", field))
			}
			if field == "timeout" {
				if value <= 0.0 || value > 10.0 {
					return NewValidationError(fmt.Sprintf("%s field is invalid. must be between 1 and 10", field))
				}
			}
			if field == "port" {
				settings[field] = int(value)
				if value < 1 || value > 65535 {
					return NewValidationError(fmt.Sprintf("%s field is invalid. must be between 1 and 65535", field))
				}
			}
		case "bool":
			_, ok := rawVal.(bool)
			if !ok {
				return NewValidationError(fmt.Sprintf("%s field is invalid type. Expected boolean", field))
			}
		case "size":
			value := int64(0)

			switch rawVal.(type) {
			case float64:
				value = int64(rawVal.(float64))
			case int64:
				value = rawVal.(int64)
			case string:
				if rawVal.(string) == "" {
					delete(c.Settings, field)
					continue
				}

				re, err := regexp.Compile(`^(?i:(\d+)([km]?)b?)$`)
				if err != nil {
					return NewValidationError(fmt.Sprintf("%s field is invalid. error compiling size regexp", field))
				}

				matched := re.FindStringSubmatch(rawVal.(string))
				if matched == nil {
					return NewValidationError(fmt.Sprintf("%s field is invalid. must be number or size string", field))
				}

				value, err = strconv.ParseInt(matched[1], 10, 64)
				if strings.ToLower(matched[2]) == "m" {
					value = value * 1024 * 1024
				} else if strings.ToLower(matched[2]) == "k" {
					value = value * 1024
				}
			default:
				return NewValidationError(fmt.Sprintf("%s field is invalid. must be number or size string", field))
			}

			if field == "downloadLimit" {
				for _, quota := range quotas {
					if quota.Target == "downloadLimit" && value > int64(quota.Limit) {
						return NewValidationError(fmt.Sprintf("%s field is invalid. %s is over limit of %s", field, formatSize(value), formatSize(int64(quota.Limit))))
					}
				}
			}
		case "ipversion":
			version, ok := rawVal.(string)
			if !ok {
				return NewValidationError(fmt.Sprintf("%s field is invalid type. Expected ip version", field))
			}
			if !(version == "v4" || version == "v6" || version == "any") {
				return NewValidationError(fmt.Sprintf("%s field is invalid. Expected v4, v6 or any", field))
			}
		}
	}
	return nil
}

func (c Check) validatePINGSettings() error {
	settings := c.Settings

	requiredFields := map[string]string{
		"hostname": "string",
	}
	optFields := map[string]string{
		"timeout":   "number",
		"ipversion": "ipversion",
	}
	for field, dataType := range requiredFields {
		rawVal, ok := settings[field]
		if !ok {
			return NewValidationError(fmt.Sprintf("%s field missing from Ping check", field))
		}
		switch dataType {
		case "string":
			value, ok := rawVal.(string)
			if !ok {
				return NewValidationError(fmt.Sprintf("%s field is invalid type. Expected string", field))
			}
			if value == "" {
				return NewValidationError(fmt.Sprintf("%s field missing from Ping check", field))
			}
		}
	}

	for field, dataType := range optFields {
		rawVal, ok := settings[field]
		if !ok {
			continue
		}
		switch dataType {
		case "number":
			value, ok := rawVal.(float64)
			if !ok {
				return NewValidationError(fmt.Sprintf("%s field is invalid type. Expected number", field))
			}
			if field == "timeout" {
				if value <= 0.0 || value > 10.0 {
					return NewValidationError(fmt.Sprintf("%s field is invalid. must be between 1 and 10", field))
				}
			}
		case "ipversion":
			version, ok := rawVal.(string)
			if !ok {
				return NewValidationError(fmt.Sprintf("%s field is invalid type. Expected ip version", field))
			}
			if !(version == "v4" || version == "v6" || version == "any") {
				return NewValidationError(fmt.Sprintf("%s field is invalid. Expected v4, v6 or any", field))
			}
		}
	}
	return nil
}

func (c Check) validateDNSSettings() error {
	settings := c.Settings

	requiredFields := map[string]string{
		"name":   "string",
		"type":   "string",
		"server": "string",
	}
	optFields := map[string]string{
		"timeout":  "float",
		"protocol": "string",
		"port":     "integer",
	}
	validRecordTypes := map[string]bool{
		"A":     true,
		"AAAA":  true,
		"CNAME": true,
		"MX":    true,
		"NS":    true,
		"PTR":   true,
		"SOA":   true,
		"SRV":   true,
		"TXT":   true,
	}
	for field, dataType := range requiredFields {
		rawVal, ok := settings[field]
		if !ok {
			return NewValidationError(fmt.Sprintf("%s field missing from DNS check", field))
		}
		switch dataType {
		case "string":
			value, ok := rawVal.(string)
			if !ok {
				return NewValidationError(fmt.Sprintf("%s field is invalid type. Expected string", field))
			}
			if value == "" {
				return NewValidationError(fmt.Sprintf("%s field missing from DNS check", field))
			}
			if field == "type" {
				if _, ok := validRecordTypes[value]; !ok {
					return NewValidationError(fmt.Sprintf("unknown dns record type: %s", value))
				}
			}
		}
	}

	for field, dataType := range optFields {
		rawVal, ok := settings[field]
		if !ok {
			continue
		}
		switch dataType {
		case "string":
			value, ok := rawVal.(string)
			if !ok {
				return NewValidationError(fmt.Sprintf("%s field is invalid type. Expected string", field))
			}
			if field == "protocol" {
				if strings.ToLower(value) != "udp" && strings.ToLower(value) != "tcp" {
					return NewValidationError("protocol field is invalid must be tcp or udp")
				}
			}
		case "number":
			value, ok := rawVal.(float64)
			if !ok {
				return NewValidationError(fmt.Sprintf("%s field is invalid type. Expected number", field))
			}
			if field == "timeout" {
				if value <= 0.0 || value > 10.0 {
					return NewValidationError(fmt.Sprintf("%s field is invalid. must be between 1 and 10", field))
				}
			}
			if field == "port" {
				settings[field] = int(value)
				if value < 1 || value > 65535 {
					return NewValidationError(fmt.Sprintf("%s field is invalid. must be between 1 and 65535", field))
				}
			}
		}
	}
	return nil
}
