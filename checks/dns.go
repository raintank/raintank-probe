package checks

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/miekg/dns"
	"github.com/raintank/raintank-probe/probe"
	m "github.com/raintank/worldping-api/pkg/models"
	"gopkg.in/raintank/schema.v1"
)

// results. we use pointers so that missing data will be
// encoded as 'null' in the json response.
type DnsResult struct {
	Time    *float64 `json:"time"`
	Ttl     *uint32  `json:"ttl"`
	Answers *int     `json:"answers"`
	Error   *string  `json:"error"`
}

func (r *DnsResult) ErrorMsg() string {
	if r.Error == nil {
		return ""
	}
	return *r.Error
}

func (r *DnsResult) Metrics(t time.Time, check *m.CheckWithSlug) []*schema.MetricData {
	metrics := make([]*schema.MetricData, 0)
	if r.Time != nil {
		metrics = append(metrics, &schema.MetricData{
			OrgId:    int(check.OrgId),
			Name:     fmt.Sprintf("worldping.%s.%s.dns.time", check.Slug, probe.Self.Slug),
			Metric:   "worldping.dns.time",
			Interval: int(check.Frequency),
			Unit:     "ms",
			Mtype:    "gauge",
			Time:     t.Unix(),
			Tags: []string{
				fmt.Sprintf("endpoint=%s", check.Slug),
				fmt.Sprintf("monitor_type=%s", check.Type),
				fmt.Sprintf("probe=%s", probe.Self.Slug),
			},
			Value: *r.Time,
		})
		metrics = append(metrics, &schema.MetricData{
			OrgId:    int(check.OrgId),
			Name:     fmt.Sprintf("worldping.%s.%s.dns.default", check.Slug, probe.Self.Slug),
			Metric:   "worldping.dns.default",
			Interval: int(check.Frequency),
			Unit:     "ms",
			Mtype:    "gauge",
			Time:     t.Unix(),
			Tags: []string{
				fmt.Sprintf("endpoint=%s", check.Slug),
				fmt.Sprintf("monitor_type=%s", check.Type),
				fmt.Sprintf("probe=%s", probe.Self.Slug),
			},
			Value: *r.Time,
		})
	}
	if r.Ttl != nil {
		metrics = append(metrics, &schema.MetricData{
			OrgId:    int(check.OrgId),
			Name:     fmt.Sprintf("worldping.%s.%s.dns.ttl", check.Slug, probe.Self.Slug),
			Metric:   "worldping.dns.ttl",
			Interval: int(check.Frequency),
			Unit:     "s",
			Mtype:    "gauge",
			Time:     t.Unix(),
			Tags: []string{
				fmt.Sprintf("endpoint=%s", check.Slug),
				fmt.Sprintf("monitor_type=%s", check.Type),
				fmt.Sprintf("probe=%s", probe.Self.Slug),
			},
			Value: float64(*r.Ttl),
		})
	}
	if r.Answers != nil {
		metrics = append(metrics, &schema.MetricData{
			OrgId:    int(check.OrgId),
			Name:     fmt.Sprintf("worldping.%s.%s.dns.answers", check.Slug, probe.Self.Slug),
			Metric:   "worldping.dns.time",
			Interval: int(check.Frequency),
			Unit:     "",
			Mtype:    "gauge",
			Time:     t.Unix(),
			Tags: []string{
				fmt.Sprintf("endpoint=%s", check.Slug),
				fmt.Sprintf("monitor_type=%s", check.Type),
				fmt.Sprintf("probe=%s", probe.Self.Slug),
			},
			Value: float64(*r.Answers),
		})
	}

	return metrics
}

type DnsRecordType string

var recordTypeToWireType = map[DnsRecordType]uint16{
	"A":     dns.TypeA,
	"AAAA":  dns.TypeAAAA,
	"CNAME": dns.TypeCNAME,
	"MX":    dns.TypeMX,
	"NS":    dns.TypeNS,
	"PTR":   dns.TypePTR,
	"SOA":   dns.TypeSOA,
	"SRV":   dns.TypeSRV,
	"TXT":   dns.TypeTXT,
}

func (t *DnsRecordType) IsValid() bool {
	_, ok := recordTypeToWireType[*t]
	return ok
}

// Our check definition.
type RaintankProbeDns struct {
	RecordName string
	RecordType DnsRecordType
	Servers    []string
	Port       int64
	Protocol   string
	Timeout    time.Duration
}

func NewRaintankDnsProbe(settings map[string]interface{}) (*RaintankProbeDns, error) {
	p := RaintankProbeDns{}
	recordName, ok := settings["name"]
	if !ok {
		return nil, fmt.Errorf("no record name passed.")
	}
	p.RecordName, ok = recordName.(string)
	if !ok {
		return nil, fmt.Errorf("invalid value for name, must be string.")
	}
	if p.RecordName == "" {
		return nil, fmt.Errorf("no record name passed.")
	}

	recordType, ok := settings["type"]
	if !ok {
		return nil, fmt.Errorf("no record type passed.")
	}
	rt, ok := recordType.(string)
	if !ok {
		return nil, fmt.Errorf("invalid value for type, must be string.")
	}
	rType := DnsRecordType(rt)
	if !rType.IsValid() {
		return nil, fmt.Errorf("invalid record type passed.")
	}

	p.RecordType = rType

	servers, ok := settings["server"]
	if !ok {
		return nil, fmt.Errorf("no servers passed.")
	}
	serverList, ok := servers.(string)
	if !ok {
		return nil, fmt.Errorf("invalid value for server, must be string.")
	}

	p.Servers = strings.Split(serverList, ",")
	if len(p.Servers) == 0 {
		return nil, fmt.Errorf("no servers passed.")
	}

	timeout, ok := settings["timeout"]
	var t float64
	if !ok {
		t = 5.0
	} else {
		t, ok = timeout.(float64)
		if !ok {
			return nil, fmt.Errorf("invalid value for timeout, must be number.")
		}
	}
	if t <= 0.0 {
		return nil, fmt.Errorf("invalid value for timeout, must be greater then 0.")
	}
	p.Timeout = time.Duration(time.Millisecond * time.Duration(int(1000.0*t)))

	port, ok := settings["port"]
	if !ok {
		p.Port = 53
	} else {
		switch port.(type) {
		case float64:
			p.Port = int64(port.(float64))
		case int64:
			p.Port = port.(int64)
		default:
			return nil, fmt.Errorf("invalid value for port, must be number.")
		}
	}
	if p.Port < 1 || p.Port > 65535 {
		return nil, fmt.Errorf("invalid port number.")
	}

	proto, ok := settings["protocol"]
	if !ok {
		p.Protocol = "udp"
	} else {
		p.Protocol, ok = proto.(string)
		if !ok {
			return nil, fmt.Errorf("invalid value for protocol, must be string.")
		}
		p.Protocol = strings.ToLower(p.Protocol)

	}
	if !(p.Protocol == "udp" || p.Protocol == "tcp") {
		return nil, fmt.Errorf("invalid protocol.")
	}

	return &p, nil
}

// run the check. this is executed in a goroutine.
func (p *RaintankProbeDns) Run() (CheckResult, error) {
	deadline := time.Now().Add(p.Timeout)
	result := &DnsResult{}
	// fix failed to respond with upper case
	c := dns.Client{Net: p.Protocol}
	m := dns.Msg{}
	if !strings.HasSuffix(p.RecordName, ".") {
		p.RecordName = p.RecordName + "."
	}
	m.SetQuestion(p.RecordName, recordTypeToWireType[p.RecordType])

	for _, s := range p.Servers {
		if time.Now().After(deadline) {
			msg := "timeout looking up dns record."
			result.Error = &msg
			return result, nil
		}
		//trim any leading/training whitespace.
		server := strings.Trim(s, " ")

		srvPort := net.JoinHostPort(server, strconv.FormatInt(p.Port, 10))
		start := time.Now()
		r, t, err := c.Exchange(&m, srvPort)
		if err != nil || r == nil {
			//try the next server.
			continue
		}
		//tcp queries dont return time.
		if t == 0 {
			t = time.Since(start)
		}
		duration := t.Seconds() * 1000
		result.Time = &duration
		answers := len(r.Answer)
		result.Answers = &answers
		if answers > 0 {
			ttl := r.Answer[0].Header().Ttl
			result.Ttl = &ttl
		}
		return result, nil
	}
	msg := "All target servers failed to respond"
	result.Error = &msg
	return result, nil
}
