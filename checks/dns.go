package checks

import (
	"encoding/json"
	"fmt"
	"github.com/miekg/dns"
	"strconv"
	"strings"
	"time"
)

// results. we use pointers so that missing data will be
// encoded as 'null' in the json response.
type DnsResult struct {
	Time    *float64 `json:"time"`
	Ttl     *uint32  `json:"ttl"`
	Answers *int     `json:"answers"`
	Error   *string  `json:"error"`
}

type DnsRecordType string

var recordTypeToWireType = map[DnsRecordType]uint16{
	"A":     dns.TypeA,
	"AAAA":  dns.TypeAAAA,
	"CNAME": dns.TypeCNAME,
	"MX":    dns.TypeMX,
	"NX":    dns.TypeNS,
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
	RecordName string        `json:"name"`
	RecordType DnsRecordType `json:"type"`
	Server     string        `json:"server"`
	Port       string        `json:"port"`
	Protocol   string        `json:"protocol"`
	Result     *DnsResult    `json:"-"`
}

// parse the json request body to build our check definition.
func NewRaintankDnsProbe(body []byte) (*RaintankProbeDns, error) {
	p := RaintankProbeDns{}
	err := json.Unmarshal(body, &p)
	if err != nil {
		return nil, fmt.Errorf("failed to parse settings. " + err.Error())
	}
	if p.Port == "" {
		p.Port = "53"
	}
	if port, err := strconv.ParseInt(p.Port, 10, 32); err != nil || port < 1 || port > 65535 {
		return nil, fmt.Errorf("failed to parse settings. Invalid port")
	}
	if p.Protocol == "" {
		p.Protocol = "udp"
	}
	p.Protocol = strings.ToLower(p.Protocol)
	if !(p.Protocol == "udp" || p.Protocol == "tcp") {
		return nil, fmt.Errorf("failed to parse settings. Invalid protocol")
	}
	return &p, nil
}

// return the results of the check
func (p *RaintankProbeDns) Results() interface{} {
	return p.Result
}

// run the check. this is executed in a goroutine.
func (p *RaintankProbeDns) Run() error {
	p.Result = &DnsResult{}
	if !p.RecordType.IsValid() {
		msg := "Invlid record type. " + string(p.RecordType)
		p.Result.Error = &msg
		return nil
	}

	servers := strings.Split(p.Server, ",")
	// fix failed to respond with upper case
	c := dns.Client{Net: p.Protocol}
	m := dns.Msg{}
	m.SetQuestion(p.RecordName+".", recordTypeToWireType[p.RecordType])

	for _, s := range servers {
		//trim any leading/training whitespace.
		server := strings.Trim(s, " ")

		srvPort := server + ":" + p.Port
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
		p.Result.Time = &duration
		answers := len(r.Answer)
		p.Result.Answers = &answers
		if answers > 0 {
			ttl := r.Answer[0].Header().Ttl
			p.Result.Ttl = &ttl
		}
		return nil
	}
	msg := "All target servers failed to respond"
	p.Result.Error = &msg
	return nil
}
