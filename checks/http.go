package checks

import (
	"encoding/json"
	"log"
	"strconv"
	"time"
	"net"
	"net/http"
	"io"
	"io/ioutil"
	"fmt"
	"strings"
	"bufio"
	"bytes"
	"regexp"
	"compress/gzip"
)

// HTTPResult struct
type HTTPResult struct {
	DNS        *float64 `json:"dns"`
	Connect    *float64 `json:"connect"`
	Send       *float64 `json:"send"`
	Wait       *float64 `json:"wait"`
	Recv       *float64 `json:"recv"`
	Total      *float64 `json:"total"`
	DataLength *float64 `json:"dataLength"`
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
func (p *RaintankProbeHTTP) Results() interface{} {
	return p.Result
}


// Run checking
func (p *RaintankProbeHTTP) Run() error {
	p.Result = &HTTPResult{}
	
	if p.Method == "" {
		p.Method = "GET"
	}
	
	// reader
	url := fmt.Sprintf("http://%s:%s%s", p.Host, p.Port, p.Path)
	request, err := http.NewRequest(p.Method, url, nil)
	request.Header.Set("Connection", "close")
	request.Header.Set("Accept-Encoding", "gzip")
	request.Header.Set("User-Agent", "raintank-probe")
	
	if err != nil {
		msg := "connection closed"
		p.Result.Error = &msg
		return nil
	}
	
	// Dialing
	start := time.Now()
	conn, err := net.Dial("tcp", p.Host+":"+p.Port)
	if err != nil {
		msg := err.Error()
		p.Result.Error = &msg
		return nil
	}
	duration := time.Since(start).Seconds() * 1000
	p.Result.Connect = &duration
	
	// DNS
	step := time.Now()
	addrs, err := net.LookupHost(p.Host)
	if err != nil || len(addrs) < 1 {
		msg := "failed to resolve hostname to IP."
		p.Result.Error = &msg
		return nil
	}
	duration = time.Since(step).Seconds() * 1000
	p.Result.DNS = &duration

	// Send
	step = time.Now()
	if err := request.Write(conn); err != nil {
		msg := err.Error()
		p.Result.Error = &msg
		return nil
	}
	send := time.Since(step).Seconds() * 1000
	p.Result.Send = &send
	defer conn.Close()
	
	// Wait & Receive
	buf := &bytes.Buffer{}
	data := make([]byte, 1)
	first := true
	step = time.Now()
	for {
		n, err := conn.Read(data)
		if err != nil {
			if err != io.EOF {
				msg := "Read error"
				p.Result.Error = &msg
				return nil
			}
			break
		}
		if first == true {
			wait := time.Since(step).Seconds() * 1000
			p.Result.Wait = &wait
			first = false
		}
		buf.Write(data[:n])
	}
	recv := time.Since(step).Seconds() * 1000
	/* 
		Total time 
	*/
	total := time.Since(start).Seconds() * 1000
	p.Result.Total = &total
	
	p.Result.Recv = &recv
	defer conn.Close()
	
	readbuffer := bytes.NewBuffer(buf.Bytes())
	response, err := http.ReadResponse(bufio.NewReader(readbuffer), request)
	
	// Error response
	statusCode := float64(response.StatusCode)
	if statusCode >= 400 {
		msg := "Invalid status code " + strconv.Itoa(response.StatusCode);
		p.Result.Error = &msg
		return nil
	}
    p.Result.StatusCode = &statusCode
	
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
	}else{
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
	
	if err != nil {
		msg := err.Error()
		p.Result.Error = &msg
		return nil
	}
	
	// Regex
	if p.ExpectRegex != "" {
		rgx, err := regexp.Compile(p.ExpectRegex)
		if err != nil {
			msg := err.Error()
			p.Result.Error = &msg
			return nil
		}
		
		if rgx.MatchString(string(body)) {
			// return ?
		}else{
			// return ?
		}
    }
	
	return nil
}
