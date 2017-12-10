package models

import (
	"errors"
	"fmt"
	"time"
)

// Typed errors
var (
	ErrMonitorNotFound          = errors.New("Monitor not found")
	ErrMonitorCollectorsInvalid = errors.New("Invalid Collector specified for Monitor")
	ErrMonitorSettingsInvalid   = errors.New("Invald variables used in Monitor Settings")
	ErrorEndpointCantBeChanged  = errors.New("A monitor's endpoint_id can not be changed.")
)

type CheckEvalResult int

const (
	EvalResultOK CheckEvalResult = iota
	EvalResultWarn
	EvalResultCrit
	EvalResultUnknown = -1
)

func (c CheckEvalResult) String() string {
	switch c {
	case EvalResultOK:
		return "OK"
	case EvalResultWarn:
		return "Warning"
	case EvalResultCrit:
		return "Critical"
	case EvalResultUnknown:
		return "Unknown"
	default:
		panic(fmt.Sprintf("Invalid CheckEvalResult value %d", int(c)))
	}
}

type MonitorType struct {
	Id      int64
	Name    string
	Created time.Time
	Updated time.Time
}

var (
	MonitorTypeToCheckTypeMap = []CheckType{
		HTTP_CHECK,
		HTTPS_CHECK,
		PING_CHECK,
		DNS_CHECK,
	}
	CheckTypeToMonitorTypeMap = map[CheckType]int64{
		HTTP_CHECK:  1,
		HTTPS_CHECK: 2,
		PING_CHECK:  3,
		DNS_CHECK:   4,
	}
)

type MonitorTypeSetting struct {
	Id            int64
	MonitorTypeId int64
	Variable      string
	Description   string
	Required      bool
	DataType      string
	DefaultValue  string
	Conditions    map[string]interface{}
}

type Monitor struct {
	Id             int64
	OrgId          int64
	EndpointId     int64
	MonitorTypeId  int64
	Offset         int64
	Frequency      int64
	Enabled        bool
	State          CheckEvalResult
	StateChange    time.Time
	StateCheck     time.Time
	Settings       []MonitorSettingDTO
	HealthSettings *CheckHealthSettings
	Created        time.Time
	Updated        time.Time
}

// ---------------
// DTOs

type MonitorForAlertDTO struct {
	Id              int64
	OrgId           int64
	EndpointId      int64
	EndpointSlug    string
	EndpointName    string
	MonitorTypeId   int64
	MonitorTypeName string
	Offset          int64
	Frequency       int64
	Enabled         bool
	StateChange     time.Time
	StateCheck      time.Time
	Settings        []MonitorSettingDTO
	HealthSettings  *CheckHealthSettings
	Created         time.Time
	Updated         time.Time
}

func (m *MonitorForAlertDTO) SettingsMap() map[string]string {
	settings := make(map[string]string)
	for _, s := range m.Settings {
		settings[s.Variable] = s.Value
	}
	return settings
}

type MonitorDTO struct {
	Id              int64                `json:"id"`
	OrgId           int64                `json:"org_id"`
	EndpointId      int64                `json:"endpoint_id" `
	EndpointSlug    string               `json:"endpoint_slug"`
	MonitorTypeId   int64                `json:"monitor_type_id"`
	MonitorTypeName string               `json:"monitor_type_name"`
	CollectorIds    []int64              `json:"collector_ids"`
	CollectorTags   []string             `json:"collector_tags"`
	Collectors      []int64              `json:"collectors"`
	State           CheckEvalResult      `json:"state"`
	StateChange     time.Time            `json:"state_change"`
	StateCheck      time.Time            `json:"state_check"`
	Settings        []MonitorSettingDTO  `json:"settings"`
	HealthSettings  *CheckHealthSettings `json:"health_settings"`
	Frequency       int64                `json:"frequency"`
	Enabled         bool                 `json:"enabled"`
	Offset          int64                `json:"offset"`
	Updated         time.Time            `json:"updated"`
}

func MonitorDTOFromCheck(c Check, endpointSlug string) MonitorDTO {
	settings := make([]MonitorSettingDTO, 0)
	for k, v := range c.Settings {
		settings = append(settings, MonitorSettingDTO{
			Variable: k,
			Value:    fmt.Sprintf("%v", v),
		})
	}

	m := MonitorDTO{
		Id:              c.Id,
		OrgId:           c.OrgId,
		EndpointId:      c.EndpointId,
		EndpointSlug:    endpointSlug,
		MonitorTypeId:   CheckTypeToMonitorTypeMap[c.Type],
		MonitorTypeName: string(c.Type),
		State:           c.State,
		StateChange:     c.StateChange,
		StateCheck:      c.StateCheck,
		HealthSettings:  c.HealthSettings,
		Settings:        settings,
		Frequency:       c.Frequency,
		Offset:          c.Offset,
		Enabled:         c.Enabled,
		Updated:         c.Updated,
	}
	if c.Route.Type == RouteByIds {
		m.CollectorIds = c.Route.Config["ids"].([]int64)
	} else if c.Route.Type == RouteByTags {
		m.CollectorTags = c.Route.Config["tags"].([]string)
	}
	return m
}

func MonitorDTOFromCheckWithSlug(c CheckWithSlug) MonitorDTO {
	settings := make([]MonitorSettingDTO, 0)
	for k, v := range c.Settings {
		settings = append(settings, MonitorSettingDTO{
			Variable: k,
			Value:    fmt.Sprintf("%v", v),
		})
	}

	m := MonitorDTO{
		Id:              c.Id,
		OrgId:           c.OrgId,
		EndpointId:      c.EndpointId,
		EndpointSlug:    c.Slug,
		MonitorTypeId:   CheckTypeToMonitorTypeMap[c.Type],
		MonitorTypeName: string(c.Type),
		State:           c.State,
		StateChange:     c.StateChange,
		StateCheck:      c.StateCheck,
		HealthSettings:  c.HealthSettings,
		Settings:        settings,
		Frequency:       c.Frequency,
		Offset:          c.Offset,
		Enabled:         c.Enabled,
		Updated:         c.Updated,
	}
	if c.Route.Type == RouteByIds {
		m.CollectorIds = c.Route.Config["ids"].([]int64)
	} else if c.Route.Type == RouteByTags {
		m.CollectorTags = c.Route.Config["tags"].([]string)
	}
	return m
}

type MonitorTypeSettingDTO struct {
	Variable     string                 `json:"variable"`
	Description  string                 `json:"description"`
	Required     bool                   `json:"required"`
	DataType     string                 `json:"data_type"`
	Conditions   map[string]interface{} `json:"conditions"`
	DefaultValue string                 `json:"default_value"`
}

type MonitorTypeDTO struct {
	Id       int64                   `json:"id"`
	Name     string                  `json:"name"`
	Settings []MonitorTypeSettingDTO `json:"settings"`
}

// ----------------------
// COMMANDS

type AddMonitorCommand struct {
	OrgId          int64                `json:"-"`
	EndpointId     int64                `json:"endpoint_id" binding:"Required"`
	MonitorTypeId  int64                `json:"monitor_type_id" binding:"Required"`
	CollectorIds   []int64              `json:"collector_ids"`
	CollectorTags  []string             `json:"collector_tags"`
	Settings       []MonitorSettingDTO  `json:"settings"`
	HealthSettings *CheckHealthSettings `json:"health_settings"`
	Frequency      int64                `json:"frequency" binding:"Required;Range(10,600)"`
	Enabled        bool                 `json:"enabled"`
	Offset         int64                `json:"-"`
	Result         *MonitorDTO          `json:"-"`
}

type UpdateMonitorCommand struct {
	Id             int64                `json:"id" binding:"Required"`
	EndpointId     int64                `json:"endpoint_id" binding:"Required"`
	OrgId          int64                `json:"-"`
	MonitorTypeId  int64                `json:"monitor_type_id" binding:"Required"`
	CollectorIds   []int64              `json:"collector_ids"`
	CollectorTags  []string             `json:"collector_tags"`
	Settings       []MonitorSettingDTO  `json:"settings"`
	HealthSettings *CheckHealthSettings `json:"health_settings"`
	Frequency      int64                `json:"frequency" binding:"Required;Range(10,600)"`
	Enabled        bool                 `json:"enabled"`
	Offset         int64                `json:"-"`
}

type DeleteMonitorCommand struct {
	Id    int64 `json:"id" binding:"Required"`
	OrgId int64 `json:"-"`
}

type UpdateMonitorStateCommand struct {
	Id       int64
	State    CheckEvalResult
	Updated  time.Time
	Checked  time.Time
	Affected int
}

// ---------------------
// QUERIES

type GetMonitorsQuery struct {
	EndpointId int64 `form:"endpoint_id"`
}
