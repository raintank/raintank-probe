package checks

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/miekg/dns"
)

// HTTPResult struct
type HTTPResult struct {
	DNS        *float64 `json:"dns"`
	Connect    *float64 `json:"connect"`
	Send       *float64 `json:"send"`
	Wait       *float64 `json:"wait"`
	Recv       *float64 `json:"recv"`
	Total      *float64 `json:"total"`
	DataLength int      `json:"dataLength"`
	StatusCode *float64 `json:"statusCode"`
	Error      *string  `json:"error"`
}

// RaintankProbeHTTP struct.
type RaintankProbeHTTP struct {
	Host        string      `json:"host"`
	Path        string      `json:"path"`
	Port        string      `json:"port"`
	Method      string      `json:"method"`
	Headers     string      `json:"headers"`
	ExpectRegex string      `json:"expectRegex"`
	Result      *HTTPResult `json:"-"`
}

// NewRaintankHTTPProbe json check
func NewRaintankHTTPProbe(body []byte) (*RaintankProbeHTTP, error) {
	p := RaintankProbeHTTP{}
	err := json.Unmarshal(body, &p)
	if err != nil {
		return nil, fmt.Errorf("failed to parse settings. " + err.Error())
	}
	if port, err := strconv.ParseInt(p.Port, 10, 32); err != nil || port < 1 || port > 65535 {
		return nil, fmt.Errorf("failed to parse settings. Invalid port")
	}
	return &p, nil
}

// Results interface
func (p *RaintankProbeHTTP) Results() interface{} {
	return p.Result
}

// Run checking
func (p *RaintankProbeHTTP) Run() error {
	p.Result = &HTTPResult{}

	request, err := http.NewRequest(p.Method, p.Host+p.Path, nil)
	if err != nil {
		msg := "connection closed" //err.Error()
		p.Result.Error = &msg
		return nil
	}

	start := time.Now()
	// Dialing
	request.Header.Set("Connection", "close")
	conn, err := net.Dial("tcp", p.Host+":"+p.Port)
	if err != nil {
		msg := err.Error()
		p.Result.Error = &msg
		return nil
	}
	defer conn.Close()
	duration := time.Since(start).Seconds()
	p.Result.Connect = &duration

	// DNS
	step := time.Now()
	server := "8.8.8.8"
	c := dns.Client{}
	m := dns.Msg{}
	m.SetQuestion(p.Host, dns.TypeA)
	_, t, err := c.Exchange(&m, server+":53")
	if err != nil {
		msg := err.Error()
		p.Result.Error = &msg
		return nil
	}
	if t == 0 {
		t = time.Since(start)
	}
	duration = t.Seconds()
	p.Result.DNS = &duration

	// Send
	step = time.Now()
	if err := request.Write(conn); err != nil {
		msg := err.Error()
		p.Result.Error = &msg
		return nil
	}
	send := time.Since(step).Seconds()
	p.Result.Send = &send

	// Wait
	step = time.Now()
	firstByte := make([]byte, 1)
	bytesRead, err := conn.Read(firstByte)
	if err != nil {
		msg := err.Error()
		p.Result.Error = &msg
		return nil
	}
	p.Result.DataLength = bytesRead
	wait := time.Since(step).Seconds()
	p.Result.Wait = &wait

	// Receive
	step = time.Now()
	var buf bytes.Buffer
	buf.Write(firstByte)
	_, err = io.Copy(&buf, conn)
	if err != nil {
		msg := err.Error()
		p.Result.Error = &msg
		return nil
	}
	recv := time.Since(step).Seconds()
	p.Result.Recv = &recv

	total := time.Since(start).Seconds()
	p.Result.Total = &total
	return nil
}
