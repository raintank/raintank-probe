package checks

import (
	//"github.com/miekg/dns"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"io"
	"time"
	"bytes"
)

type HttpResult struct {
	Dns        *float64 `json:"dns"`
	Connect    *float64 `json:"connect"`
	Send       *float64 `json:"send"`
	Wait       *float64 `json:"wait"`
	Recv       *float64 `json:"recv"`
	Total      *float64 `json:"total"`
	DataLength      int `json:"dataLength"`
	StatusCode      *float64 `json:"statusCode"`
	// Connect    *time.Duration `json:"connect"`
	// Send       *time.Duration `json:"send"`
	// Wait       *time.Duration `json:"wait"`
	// Recv       *time.Duration `json:"recv"`
	// Total      *time.Duration `json:"total"`
	Error      *string  `json:"error"`
}

// Our check definition.
type RaintankProbeHttp struct {
	Host string         `json:"host"`
    Path string         `json:"path"`
    Port string         `json:"port"`
    Method string       `json:"method"`
    Headers string      `json:"headers"`
    ExpectRegex string  `json:"expectRegex"`
	Result *HttpResult  `json:"-"`
}

func NewRaintankHttpProbe(body []byte) (*RaintankProbeHttp, error) {
	p := RaintankProbeHttp{}
	err := json.Unmarshal(body, &p)
	if err != nil {
		return nil, fmt.Errorf("failed to parse settings. " + err.Error())
	}
	if port, err := strconv.ParseInt(p.Port, 10, 32); err != nil || port < 1 || port > 65535 {
		return nil, fmt.Errorf("failed to parse settings. Invalid port")
	}
	return &p, nil
}

func (p *RaintankProbeHttp) Results() interface{} {
	return p.Result
}

func (p *RaintankProbeHttp) Run() error {
    p.Result = &HttpResult{}
	fmt.Println("connect")
    request, err := http.NewRequest(p.Method, p.Host + p.Path, nil)
	if err != nil {
		msg := "connection closed" //err.Error()
		p.Result.Error = &msg
		return nil
	}
	request.Header.Set("Connection", "close")
	start := time.Now()

	// Connect
	conn, err := net.Dial("tcp", p.Host + ":" + p.Port)
	if err != nil {
		msg := err.Error()
		p.Result.Error = &msg
		return nil
	}
	defer conn.Close()
	result := float64(time.Since(start))
	p.Result.Connect = &result

	// Send
	step := time.Now()
	if err := request.Write(conn); err != nil {
		msg := err.Error()
		p.Result.Error = &msg
		return nil
	}
	send := float64(time.Since(step))
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
	wait := float64(time.Since(step))
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
	recv := float64(time.Since(step))
	p.Result.Recv = &recv

	total := float64(time.Since(start))
	p.Result.Total = &total
	return nil
}
