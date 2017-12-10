package models

import (
	"regexp"
	"strings"
	"time"
)

// Typed errors
var (
	ErrProbeNotFound           = NewNotFoundError("Probe not found")
	ErrProbeWithSameCodeExists = NewValidationError("A Probe with the same code already exists")
)

type Probe struct {
	Id            int64
	OrgId         int64
	Slug          string
	Name          string
	Public        bool
	Latitude      float64
	Longitude     float64
	Created       time.Time
	Updated       time.Time
	Online        bool
	OnlineChange  time.Time
	Enabled       bool
	EnabledChange time.Time
}

type ProbeTag struct {
	Id      int64
	OrgId   int64
	ProbeId int64
	Tag     string
	Created time.Time
}

type ProbeSession struct {
	Id         int64
	OrgId      int64
	ProbeId    int64
	SocketId   string
	Version    string
	InstanceId string
	RemoteIp   string
	Updated    time.Time
}

// ----------------------
// DTO
type ProbeDTO struct {
	Id            int64     `json:"id" binding:"required"`
	OrgId         int64     `json:"org_id"`
	Slug          string    `json:"slug"`
	Name          string    `json:"name" binding:"required"`
	Tags          []string  `json:"tags"`
	Public        bool      `json:"public"`
	Latitude      float64   `json:"latitude"`
	Longitude     float64   `json:"longitude"`
	Online        bool      `json:"online"`
	OnlineChange  time.Time `json:"online_change"`
	Enabled       bool      `json:"enabled"`
	EnabledChange time.Time `json:"enabled_change"`
	Created       time.Time `json:"created"`
	Updated       time.Time `json:"updated"`
	RemoteIp      []string  `json:"remoteIp"`
}

type ProbeLocationDTO struct {
	Key       string  `json:"key"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Name      string  `json:"name"`
}

type ProbeReadyPayload struct {
	Collector    *ProbeDTO        `json:"collector"`
	MonitorTypes []MonitorTypeDTO `json:"monitor_types"`
	SocketId     string           `json:"socket_id"`
}

// ---------------------
// QUERIES

type GetProbesQuery struct {
	OrgId   int64  `form:"-"`
	Public  string `form:"public"`
	Enabled string `form:"enabled"`
	Online  string `form:"online"`
	Name    string `form:"name"`
	Slug    string `form:"slug"`
	Tag     string `form:"tag"`
	OrderBy string `form:"orderBy" binding:"In(name,slug,created,updated,)"`
}

func (collector *Probe) UpdateSlug() {
	name := strings.ToLower(collector.Name)
	re := regexp.MustCompile("[^\\w ]+")
	re2 := regexp.MustCompile("\\s")
	collector.Slug = re2.ReplaceAllString(re.ReplaceAllString(name, ""), "-")
}
