package checks

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// HTTPSResult struct
type HTTPSResult struct {
	DNS        *float64 `json:"dns"`
	Connect    *float64 `json:"connect"`
	Send       *float64 `json:"send"`
	Wait       *float64 `json:"wait"`
	Recv       *float64 `json:"recv"`
	Total      *float64 `json:"total"`
	DataLength *float64 `json:"dataLength"`
	StatusCode *float64 `json:"statusCode"`
	Expiry     *string  `json:"expiry"`
	Error      *string  `json:"error"`
}

// RaintankProbeHTTPS struct.
type RaintankProbeHTTPS struct {
	Host         string       `json:"host"`
	Path         string       `json:"path"`
	Port         string       `json:"port"`
	ValidateCert string       `json:"validateCert"`
	Method       string       `json:"method"`
	Headers      string       `json:"headers"`
	ExpectRegex  string       `json:"expectRegex"`
	Result       *HTTPSResult `json:"-"`
}

// NewRaintankHTTPSProbe json check
func NewRaintankHTTPSProbe(body []byte) (*RaintankProbeHTTPS, error) {
	p := RaintankProbeHTTPS{}
	err := json.Unmarshal(body, &p)
	if err != nil {
		log.Fatalf("failed to parse settings. %v", err.Error())
		return nil, err
	}
	if port, err := strconv.ParseInt(p.Port, 10, 32); err != nil || port < 1 || port > 65535 {
		log.Fatal("failed to parse settings. Invalid port")
		return nil, err
	}
	if !strings.HasPrefix(p.Path, "/") {
		return nil, fmt.Errorf("invalid path. must start with /")
	}
	return &p, nil
}

// Results interface
func (p *RaintankProbeHTTPS) Results() interface{} {
	return p.Result
}

// Run checking
func (p *RaintankProbeHTTPS) Run() error {
	p.Result = &HTTPSResult{}

	if p.Method == "" {
		p.Method = "GET"
	}

	// reader
	url := fmt.Sprintf("http://%s:%s%s", p.Host, p.Port, p.Path)
	request, err := http.NewRequest(p.Method, url, nil)

	// Parsing header (use fake request)
	if p.Headers != "" {
		headReader := bufio.NewReader(strings.NewReader("GET / HTTP/1.1\r\n" + p.Headers + "\r\n\r\n"))
		dummyRequest, err := http.ReadRequest(headReader)
		if err != nil {
			msg := err.Error()
			p.Result.Error = &msg
			return nil
		}

		for key := range dummyRequest.Header {
			request.Header.Set(key, dummyRequest.Header.Get(key))
		}
	}

	if _, found := request.Header["Accept-Encoding"]; !found {
		request.Header.Set("Accept-Encoding", "gzip")
	}

	if err != nil {
		msg := "connection closed"
		p.Result.Error = &msg
		return nil
	}

	// DNS lookup
	step := time.Now()
	addrs, err := net.LookupHost(p.Host)
	if err != nil || len(addrs) < 1 {
		msg := "failed to resolve hostname to IP."
		p.Result.Error = &msg
		return nil
	}
	dnsResolve := time.Since(step).Seconds() * 1000
	p.Result.DNS = &dnsResolve

	validate, _ := strconv.ParseBool(p.ValidateCert)
	tlsConfig := &tls.Config{
		InsecureSkipVerify: !validate,
		ServerName:         p.Host,
	}

	// Dialing
	start := time.Now()
	conn, err := tls.Dial("tcp", addrs[0]+":"+p.Port, tlsConfig)
	if err != nil {
		msg := err.Error()
		p.Result.Error = &msg
		return nil
	}
	defer conn.Close()
	connecting := time.Since(start).Seconds() * 1000
	p.Result.Connect = &connecting

	// Send
	step = time.Now()
	if err := request.Write(conn); err != nil {
		msg := err.Error()
		p.Result.Error = &msg
		return nil
	}
	send := time.Since(step).Seconds() * 1000
	p.Result.Send = &send

	// Wait & Receive
	step = time.Now()

	// read first byte data
	firstData := make([]byte, 1)
	_, err = conn.Read(firstData)
	if err != nil {
		if err != io.EOF {
			msg := "Read error"
			p.Result.Error = &msg
			return nil
		}
	}
	wait := time.Since(step).Seconds() * 1000
	p.Result.Wait = &wait

	var buf bytes.Buffer
	buf.Write(firstData)
	data := make([]byte, 1024)
	limit := 100 * 1024
	for {
		n, err := conn.Read(data)
		if err != nil {
			if err != io.EOF {
				msg := "Read error"
				p.Result.Error = &msg
				return nil
			} else {
				break
			}
		}
		buf.Write(data[:n])
		if buf.Len() > limit {
			conn.Close()
			break
		}
	}

	recv := time.Since(step).Seconds() * 1000
	total := time.Since(start).Seconds() * 1000
	p.Result.Total = &total
	p.Result.Recv = &recv

	readbuffer := bytes.NewBuffer(buf.Bytes())
	response, err := http.ReadResponse(bufio.NewReader(readbuffer), request)

	if err != nil {
		msg := err.Error()
		p.Result.Error = &msg
		return nil
	}

	var reader io.ReadCloser
	if p.ExpectRegex != "" {
		// Handle gzip
		switch response.Header.Get("Content-Encoding") {
		case "gzip":
			reader, err = gzip.NewReader(response.Body)
			defer reader.Close()
			if err != nil {
				msg := err.Error()
				p.Result.Error = &msg
				return nil
			}
		default:
			reader = response.Body
		}
	} else {
		reader = response.Body
	}

	defer response.Body.Close()
	body, err := ioutil.ReadAll(reader)
	if err != nil {
		msg := err.Error()
		p.Result.Error = &msg
		return nil
	}

	// Data Length
	dataLength := float64(0)
	if dataLength, err = strconv.ParseFloat(response.Header.Get("Content-Length"), 64); dataLength < 1 || err != nil {
		dataLength = float64(len(body))
	}
	p.Result.DataLength = &dataLength

	// Error response
	statusCode := float64(response.StatusCode)
	p.Result.StatusCode = &statusCode
	if statusCode >= 400 {
		msg := "Invalid status code " + strconv.Itoa(response.StatusCode)
		p.Result.Error = &msg
		return nil
	}

	// Read certificate
	certs := conn.ConnectionState().PeerCertificates

	if certs == nil || len(certs) < 1 {
		msg := "response has no TLS field"
		p.Result.Error = &msg
	} else {
		msg := fmt.Sprintf("Subject: %s - Expires: %s\n", certs[0].Subject.CommonName, certs[0].NotAfter)
		p.Result.Expiry = &msg
	}

	// Regex
	if p.ExpectRegex != "" {
		rgx, err := regexp.Compile(p.ExpectRegex)
		if err != nil {
			msg := err.Error()
			p.Result.Error = &msg
			return nil
		}

		if !rgx.MatchString(string(body)) {
			msg := "expectRegex did not match"
			p.Result.Error = &msg
			return nil
		}
	}

	return nil
}

func ExpiresIn(t time.Time) string {
	units := [...]struct {
		suffix string
		unit   time.Duration
	}{
		{"days", 24 * time.Hour},
		{"hours", time.Hour},
		{"minutes", time.Minute},
		{"seconds", time.Second},
	}
	d := t.Sub(time.Now())
	for _, u := range units {
		if d > u.unit {
			return fmt.Sprintf("Expires in %d %s", d/u.unit, u.suffix)
		}
	}
	return fmt.Sprintf("Expired on %s", t.Local())
}
