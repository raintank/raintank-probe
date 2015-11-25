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
	Throughput *float64 `json:"throughput"`
	StatusCode *float64 `json:"statusCode"`
	Expiry     *float64 `json:"expiry"`
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
	Timeout      int          `json:"timeout"`
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

	return &p, nil
}

// Results interface
func (p *RaintankProbeHTTPS) Results() interface{} {
	return p.Result
}

// Run checking
func (p *RaintankProbeHTTPS) Run() error {
	deadline := time.Now().Add(time.Second * time.Duration(p.Timeout))
	p.Result = &HTTPSResult{}

	if port, err := strconv.ParseInt(p.Port, 10, 32); err != nil || port < 1 || port > 65535 {
		msg := "Invalid port"
		p.Result.Error = &msg
		return nil
	}
	if !strings.HasPrefix(p.Path, "/") {
		msg := "Invalid path. must start with /"
		p.Result.Error = &msg
		return nil
	}

	if p.Method == "" {
		p.Method = "GET"
	}

	// reader
	tmpPort := ""
	if p.Port != "443" {
		tmpPort = ":" + p.Port
	}
	url := fmt.Sprintf("https://%s%s%s", p.Host, tmpPort, p.Path)
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

	//always close the conneciton
	request.Header.Set("Connection", "close")

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
	tcpconn, err := net.DialTimeout("tcp", addrs[0]+":"+p.Port, 10*time.Second)
	if err != nil {
		msg := err.Error()
		p.Result.Error = &msg
		return nil
	}
	tcpconn.SetDeadline(deadline)
	conn := tls.Client(tcpconn, tlsConfig)
	err = conn.Handshake()
	if err != nil {
		msg := err.Error()
		p.Result.Error = &msg
		return nil
	}
	// Read certificate
	certs := conn.ConnectionState().PeerCertificates

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

	// Wait
	step = time.Now()

	//read first byte data
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

	//Start recieve
	step = time.Now()

	var buf bytes.Buffer
	buf.Write(firstData)

	limit := 100 * 1024
	for {
		data := make([]byte, 1024)
		n, err := conn.Read(data)
		buf.Write(data[:n])
		if err != nil {
			if err != io.EOF {
				msg := "Read error"
				p.Result.Error = &msg
				return nil
			} else {
				break
			}
		}

		if buf.Len() >= limit {
			conn.Close()
			break
		}
	}

	recv := time.Since(step).Seconds() * 1000
	total := time.Since(start).Seconds() * 1000
	p.Result.Total = &total
	p.Result.Recv = &recv

	// Data Length
	dataLength := float64(buf.Len())
	p.Result.DataLength = &dataLength

	//throughput
	if recv > 0 && dataLength > 0 {
		throughput := dataLength / (recv / 1000.0)
		p.Result.Throughput = &throughput
	}

	response, err := http.ReadResponse(bufio.NewReader(&buf), request)

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
			if err != nil {
				log.Printf("failed to decode content body for request to %s. %s", url, err)
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

	body, err := ioutil.ReadAll(reader)
	reader.Close()
	if err != nil && len(body) == 0 {
		msg := err.Error()
		p.Result.Error = &msg
		return nil
	}

	// Error response
	statusCode := float64(response.StatusCode)
	p.Result.StatusCode = &statusCode
	if statusCode >= 400 {
		msg := "Invalid status code " + strconv.Itoa(response.StatusCode)
		p.Result.Error = &msg
		return nil
	}

	if certs == nil || len(certs) < 1 {
		log.Printf("no PeerCerticates for connection to %s", p.Host)
	} else {
		timeTilExpiry := certs[0].NotAfter.Sub(time.Now())
		secondsTilExpiry := float64(timeTilExpiry) / float64(time.Second)
		p.Result.Expiry = &secondsTilExpiry
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
