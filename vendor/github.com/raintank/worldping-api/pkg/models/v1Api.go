package models

import (
	"strconv"
)

type AddEndpointCommand struct {
	OrgId    int64                `json:"-"`
	Name     string               `json:"name" binding:"Required"`
	Tags     []string             `json:"tags"`
	Monitors []*AddMonitorCommand `json:"monitors"`
	Result   *EndpointDTO         `json:"-"`
}

type UpdateEndpointCommand struct {
	Id     int64    `json:"id" binding:"Required"`
	OrgId  int64    `json:"-"`
	Name   string   `json:"name" binding:"Required"`
	Tags   []string `json:"tags"`
	Result *EndpointDTO
}

type SuggestedMonitor struct {
	MonitorTypeId int64               `json:"monitor_type_id"`
	Settings      []MonitorSettingDTO `json:"settings"`
}

type MonitorSettingDTO struct {
	Variable string `json:"variable" binding:"Required"`
	Value    string `json:"value"`
}

type MonitorSettingsDTO []MonitorSettingDTO

func (s MonitorSettingsDTO) ToV2Setting(t CheckType) map[string]interface{} {
	settings := make(map[string]interface{})
	switch t {
	case HTTP_CHECK:
		for _, v := range s {
			switch v.Variable {
			case "host":
				settings["host"] = v.Value
			case "path":
				settings["path"] = v.Value
			case "port":
				settings["port"], _ = strconv.ParseInt(v.Value, 10, 64)
			case "method":
				settings["method"] = v.Value
			case "headers":
				settings["headers"] = v.Value
			case "expectRegex":
				settings["expectRegex"] = v.Value
			case "timeout":
				settings["timeout"], _ = strconv.ParseFloat(v.Value, 64)
			}
		}
	case HTTPS_CHECK:
		for _, v := range s {
			switch v.Variable {
			case "host":
				settings["host"] = v.Value
			case "path":
				settings["path"] = v.Value
			case "port":
				settings["port"], _ = strconv.ParseInt(v.Value, 10, 64)
			case "method":
				settings["method"] = v.Value
			case "headers":
				settings["headers"] = v.Value
			case "validateCert":
				settings["validateCert"], _ = strconv.ParseBool(v.Value)
			case "expectRegex":
				settings["expectRegex"] = v.Value
			case "timeout":
				settings["timeout"], _ = strconv.ParseFloat(v.Value, 64)
			}
		}
	case PING_CHECK:
		for _, v := range s {
			switch v.Variable {
			case "hostname":
				settings["hostname"] = v.Value
			case "timeout":
				settings["timeout"], _ = strconv.ParseFloat(v.Value, 64)
			}
		}
	case DNS_CHECK:
		for _, v := range s {
			switch v.Variable {
			case "name":
				settings["name"] = v.Value
			case "type":
				settings["type"] = v.Value
			case "server":
				settings["server"] = v.Value
			case "port":
				settings["port"], _ = strconv.ParseInt(v.Value, 10, 64)
			case "protocol":
				settings["protocol"] = v.Value
			case "timeout":
				settings["timeout"], _ = strconv.ParseFloat(v.Value, 64)
			}
		}
	}
	return settings
}

var MonitorTypes = []MonitorTypeDTO{
	{
		Id:   1,
		Name: "HTTP",
		Settings: []MonitorTypeSettingDTO{
			{
				Variable:     "host",
				Description:  "Hostname",
				Required:     true,
				DataType:     "String",
				Conditions:   map[string]interface{}{},
				DefaultValue: "",
			},
			{
				Variable:     "path",
				Description:  "Path",
				Required:     true,
				DataType:     "String",
				Conditions:   map[string]interface{}{},
				DefaultValue: "/",
			},
			{
				Variable:     "port",
				Description:  "Port",
				Required:     false,
				DataType:     "Number",
				Conditions:   map[string]interface{}{},
				DefaultValue: "80",
			},
			{
				Variable:    "method",
				Description: "Method",
				Required:    false,
				DataType:    "Enum",
				Conditions: map[string]interface{}{
					"values": []string{"GET", "POST", "PUT", "DELETE", "HEAD"},
				},
				DefaultValue: "GET",
			},
			{
				Variable:     "headers",
				Description:  "Headers",
				Required:     false,
				DataType:     "Text",
				Conditions:   map[string]interface{}{},
				DefaultValue: "Accept-Encoding: gzip\nUser-Agent: raintank collector\n",
			},
			{
				Variable:     "expectRegex",
				Description:  "Content Match",
				Required:     false,
				DataType:     "String",
				Conditions:   map[string]interface{}{},
				DefaultValue: "",
			},
			{
				Variable:     "timeout",
				Description:  "Timeout",
				Required:     true,
				DataType:     "Number",
				Conditions:   map[string]interface{}{},
				DefaultValue: "5",
			},
		},
	},
	{
		Id:   2,
		Name: "HTTPS",
		Settings: []MonitorTypeSettingDTO{
			{
				Variable:     "host",
				Description:  "Hostname",
				Required:     true,
				DataType:     "String",
				Conditions:   map[string]interface{}{},
				DefaultValue: "",
			},
			{
				Variable:     "path",
				Description:  "Path",
				Required:     true,
				DataType:     "String",
				Conditions:   map[string]interface{}{},
				DefaultValue: "/",
			},
			{
				Variable:     "port",
				Description:  "Port",
				Required:     false,
				DataType:     "Number",
				Conditions:   map[string]interface{}{},
				DefaultValue: "443",
			},
			{
				Variable:    "method",
				Description: "Method",
				Required:    false,
				DataType:    "Enum",
				Conditions: map[string]interface{}{
					"values": []string{"GET", "POST", "PUT", "DELETE", "HEAD"},
				},
				DefaultValue: "GET",
			},
			{
				Variable:     "headers",
				Description:  "Headers",
				Required:     false,
				DataType:     "Text",
				Conditions:   map[string]interface{}{},
				DefaultValue: "Accept-Encoding: gzip\nUser-Agent: raintank collector\n",
			},
			{
				Variable:     "expectRegex",
				Description:  "Content Match",
				Required:     false,
				DataType:     "String",
				Conditions:   map[string]interface{}{},
				DefaultValue: "",
			},
			{
				Variable:     "validateCert",
				Description:  "Validate SSL Certificate",
				Required:     false,
				DataType:     "Boolean",
				Conditions:   map[string]interface{}{},
				DefaultValue: "true",
			},
			{
				Variable:     "timeout",
				Description:  "Timeout",
				Required:     true,
				DataType:     "Number",
				Conditions:   map[string]interface{}{},
				DefaultValue: "5",
			},
		},
	},
	{
		Id:   3,
		Name: "Ping",
		Settings: []MonitorTypeSettingDTO{
			{
				Variable:     "hostname",
				Description:  "Hostname",
				Required:     true,
				DataType:     "String",
				Conditions:   map[string]interface{}{},
				DefaultValue: "",
			},
			{
				Variable:     "timeout",
				Description:  "Timeout",
				Required:     true,
				DataType:     "Number",
				Conditions:   map[string]interface{}{},
				DefaultValue: "5",
			},
		},
	},
	{
		Id:   4,
		Name: "DNS",
		Settings: []MonitorTypeSettingDTO{
			{
				Variable:     "name",
				Description:  "Record Name",
				Required:     true,
				DataType:     "String",
				Conditions:   map[string]interface{}{},
				DefaultValue: "",
			},
			{
				Variable:    "type",
				Description: "Record Type",
				Required:    true,
				DataType:    "Enum",
				Conditions: map[string]interface{}{
					"values": []string{"A", "AAAA", "CNAME", "MX", "NS", "PTR", "SOA", "SRV", "TXT"},
				},
				DefaultValue: "A",
			},
			{
				Variable:     "server",
				Description:  "Server",
				Required:     true,
				DataType:     "String",
				Conditions:   map[string]interface{}{},
				DefaultValue: "",
			},
			{
				Variable:     "port",
				Description:  "Port",
				Required:     false,
				DataType:     "Number",
				Conditions:   map[string]interface{}{},
				DefaultValue: "53",
			},
			{
				Variable:    "protocol",
				Description: "Protocol",
				Required:    false,
				DataType:    "Enum",
				Conditions: map[string]interface{}{
					"values": []string{"tcp", "udp"},
				},
				DefaultValue: "udp",
			},
			{
				Variable:     "timeout",
				Description:  "Timeout",
				Required:     true,
				DataType:     "Number",
				Conditions:   map[string]interface{}{},
				DefaultValue: "5",
			},
		},
	},
}
