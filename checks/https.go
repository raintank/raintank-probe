package checks

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"crypto/tls"
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

	"github.com/raintank/raintank-probe/probe"
	m "github.com/raintank/worldping-api/pkg/models"
	"gopkg.in/raintank/schema.v1"
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

func (r *HTTPSResult) ErrorMsg() string {
	if r.Error == nil {
		return ""
	}
	return *r.Error
}

func (r *HTTPSResult) Metrics(t time.Time, check *m.CheckWithSlug) []*schema.MetricData {
	metrics := make([]*schema.MetricData, 0)
	if r.DNS != nil {
		metrics = append(metrics, &schema.MetricData{
			OrgId:    int(check.OrgId),
			Name:     fmt.Sprintf("worldping.%s.%s.https.dns", check.Slug, probe.Self.Slug),
			Metric:   "worldping.https.dns",
			Interval: int(check.Frequency),
			Unit:     "ms",
			Mtype:    "gauge",
			Time:     t.Unix(),
			Tags: []string{
				fmt.Sprintf("endpoint:%s", check.Slug),
				fmt.Sprintf("monitor_type:%s", check.Type),
				fmt.Sprintf("probe:%s", probe.Self.Slug),
			},
			Value: *r.DNS,
		})
	}
	if r.Connect != nil {
		metrics = append(metrics, &schema.MetricData{
			OrgId:    int(check.OrgId),
			Name:     fmt.Sprintf("worldping.%s.%s.https.connect", check.Slug, probe.Self.Slug),
			Metric:   "worldping.https.connect",
			Interval: int(check.Frequency),
			Unit:     "ms",
			Mtype:    "gauge",
			Time:     t.Unix(),
			Tags: []string{
				fmt.Sprintf("endpoint:%s", check.Slug),
				fmt.Sprintf("monitor_type:%s", check.Type),
				fmt.Sprintf("probe:%s", probe.Self.Slug),
			},
			Value: *r.Connect,
		})
	}
	if r.Send != nil {
		metrics = append(metrics, &schema.MetricData{
			OrgId:    int(check.OrgId),
			Name:     fmt.Sprintf("worldping.%s.%s.https.send", check.Slug, probe.Self.Slug),
			Metric:   "worldping.https.send",
			Interval: int(check.Frequency),
			Unit:     "ms",
			Mtype:    "gauge",
			Time:     t.Unix(),
			Tags: []string{
				fmt.Sprintf("endpoint:%s", check.Slug),
				fmt.Sprintf("monitor_type:%s", check.Type),
				fmt.Sprintf("probe:%s", probe.Self.Slug),
			},
			Value: *r.Send,
		})
	}
	if r.Wait != nil {
		metrics = append(metrics, &schema.MetricData{
			OrgId:    int(check.OrgId),
			Name:     fmt.Sprintf("worldping.%s.%s.https.wait", check.Slug, probe.Self.Slug),
			Metric:   "worldping.https.wait",
			Interval: int(check.Frequency),
			Unit:     "ms",
			Mtype:    "gauge",
			Time:     t.Unix(),
			Tags: []string{
				fmt.Sprintf("endpoint:%s", check.Slug),
				fmt.Sprintf("monitor_type:%s", check.Type),
				fmt.Sprintf("probe:%s", probe.Self.Slug),
			},
			Value: *r.Wait,
		})
	}
	if r.Recv != nil {
		metrics = append(metrics, &schema.MetricData{
			OrgId:    int(check.OrgId),
			Name:     fmt.Sprintf("worldping.%s.%s.https.recv", check.Slug, probe.Self.Slug),
			Metric:   "worldping.https.recv",
			Interval: int(check.Frequency),
			Unit:     "ms",
			Mtype:    "gauge",
			Time:     t.Unix(),
			Tags: []string{
				fmt.Sprintf("endpoint:%s", check.Slug),
				fmt.Sprintf("monitor_type:%s", check.Type),
				fmt.Sprintf("probe:%s", probe.Self.Slug),
			},
			Value: *r.Recv,
		})
	}
	if r.Total != nil {
		metrics = append(metrics, &schema.MetricData{
			OrgId:    int(check.OrgId),
			Name:     fmt.Sprintf("worldping.%s.%s.https.total", check.Slug, probe.Self.Slug),
			Metric:   "worldping.https.total",
			Interval: int(check.Frequency),
			Unit:     "ms",
			Mtype:    "gauge",
			Time:     t.Unix(),
			Tags: []string{
				fmt.Sprintf("endpoint:%s", check.Slug),
				fmt.Sprintf("monitor_type:%s", check.Type),
				fmt.Sprintf("probe:%s", probe.Self.Slug),
			},
			Value: *r.Total,
		})
		metrics = append(metrics, &schema.MetricData{
			OrgId:    int(check.OrgId),
			Name:     fmt.Sprintf("worldping.%s.%s.https.default", check.Slug, probe.Self.Slug),
			Metric:   "worldping.https.default",
			Interval: int(check.Frequency),
			Unit:     "ms",
			Mtype:    "gauge",
			Time:     t.Unix(),
			Tags: []string{
				fmt.Sprintf("endpoint:%s", check.Slug),
				fmt.Sprintf("monitor_type:%s", check.Type),
				fmt.Sprintf("probe:%s", probe.Self.Slug),
			},
			Value: *r.Total,
		})
	}
	if r.Throughput != nil {
		metrics = append(metrics, &schema.MetricData{
			OrgId:    int(check.OrgId),
			Name:     fmt.Sprintf("worldping.%s.%s.https.throughput", check.Slug, probe.Self.Slug),
			Metric:   "worldping.https.throughput",
			Interval: int(check.Frequency),
			Unit:     "B/s",
			Mtype:    "rate",
			Time:     t.Unix(),
			Tags: []string{
				fmt.Sprintf("endpoint:%s", check.Slug),
				fmt.Sprintf("monitor_type:%s", check.Type),
				fmt.Sprintf("probe:%s", probe.Self.Slug),
			},
			Value: *r.Throughput,
		})
	}
	if r.DataLength != nil {
		metrics = append(metrics, &schema.MetricData{
			OrgId:    int(check.OrgId),
			Name:     fmt.Sprintf("worldping.%s.%s.https.dataLength", check.Slug, probe.Self.Slug),
			Metric:   "worldping.https.dataLength",
			Interval: int(check.Frequency),
			Unit:     "B",
			Mtype:    "gauge",
			Time:     t.Unix(),
			Tags: []string{
				fmt.Sprintf("endpoint:%s", check.Slug),
				fmt.Sprintf("monitor_type:%s", check.Type),
				fmt.Sprintf("probe:%s", probe.Self.Slug),
			},
			Value: *r.DataLength,
		})
	}
	if r.StatusCode != nil {
		metrics = append(metrics, &schema.MetricData{
			OrgId:    int(check.OrgId),
			Name:     fmt.Sprintf("worldping.%s.%s.https.statusCode", check.Slug, probe.Self.Slug),
			Metric:   "worldping.https.statusCode",
			Interval: int(check.Frequency),
			Unit:     "",
			Mtype:    "gauge",
			Time:     t.Unix(),
			Tags: []string{
				fmt.Sprintf("endpoint:%s", check.Slug),
				fmt.Sprintf("monitor_type:%s", check.Type),
				fmt.Sprintf("probe:%s", probe.Self.Slug),
			},
			Value: *r.StatusCode,
		})
	}
	return metrics
}

// RaintankProbeHTTPS struct.
type RaintankProbeHTTPS struct {
	Host         string        `json:"host"`
	Path         string        `json:"path"`
	Port         int64         `json:"port"`
	ValidateCert bool          `json:"validateCert"`
	Method       string        `json:"method"`
	Headers      string        `json:"headers"`
	ExpectRegex  string        `json:"expectRegex"`
	Body         string        `json:"body"`
	Timeout      time.Duration `json:"timeout"`
	DownloadLimit int64         `json:"downloadLimit"`
}

// NewRaintankHTTPSProbe json check
func NewRaintankHTTPSProbe(settings map[string]interface{}) (*RaintankProbeHTTPS, error) {
	p := RaintankProbeHTTPS{}
	host, ok := settings["host"]
	if !ok {
		return nil, fmt.Errorf("no host passed.")
	}
	p.Host, ok = host.(string)
	if !ok {
		return nil, fmt.Errorf("invalid value for host, must be string.")
	}
	if p.Host == "" {
		return nil, fmt.Errorf("no host passed.")
	}

	path, ok := settings["path"]
	if !ok {
		return nil, fmt.Errorf("no path passed.")
	}
	p.Path, ok = path.(string)
	if !ok {
		return nil, fmt.Errorf("invalid value for path, must be string.")
	}
	if p.Path == "" {
		p.Path = "/"
	}

	method, ok := settings["method"]
	if !ok {
		p.Method = "GET"
	} else {
		p.Method, ok = method.(string)
		if !ok {
			return nil, fmt.Errorf("invalid value for method, must be string.")
		}
	}

	headers, ok := settings["headers"]
	if !ok {
		p.Headers = ""
	} else {
		p.Headers, ok = headers.(string)
		if !ok {
			return nil, fmt.Errorf("invalid value for headers, must be string.")
		}
	}

	expectRegex, ok := settings["expectRegex"]
	if !ok {
		p.ExpectRegex = ""
	} else {
		p.ExpectRegex, ok = expectRegex.(string)
		if !ok {
			return nil, fmt.Errorf("invalid value for expectRegex, must be string.")
		}
	}

	validateCert, ok := settings["validateCert"]
	if !ok {
		p.ValidateCert = true
	} else {
		p.ValidateCert, ok = validateCert.(bool)
		if !ok {
			return nil, fmt.Errorf("invalid value for validateCert, must be boolean.")
		}
	}

	body, ok := settings["body"]
	if !ok {
		p.Body = ""
	} else {
		p.Body, ok = body.(string)
		if !ok {
			return nil, fmt.Errorf("invalid value for body, must be string.")
		}
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
		p.Port = 443
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

	limit, ok := settings["downloadLimit"]
	if !ok {
		p.DownloadLimit = 100 * 1024
	} else {
		switch limit.(type) {
		case float64:
			p.DownloadLimit = int64(limit.(float64))
		case int64:
			p.DownloadLimit = limit.(int64)
		case string:
			re, err := regexp.Compile(`^(?i:(\d+)([km]?)b?)$`)
			if err != nil {
				return nil, fmt.Errorf("error compiling download limit regexp")
			}

			matched := re.FindStringSubmatch(limit.(string))
			if matched == nil {
				return nil, fmt.Errorf("invalid value for downloadLimit, must be number or size string.")
			}

			p.DownloadLimit, err = strconv.ParseInt(matched[1], 10, 64)
			if strings.ToLower(matched[2]) == "m" {
				p.DownloadLimit = p.DownloadLimit * 1024 * 1024
			} else if strings.ToLower(matched[2]) == "k" {
				p.DownloadLimit = p.DownloadLimit * 1024
			}
		default:
			return nil, fmt.Errorf("invalid value for downloadLimit, must be number or size string.")
		}
	}

	return &p, nil
}

// Run checking
func (p *RaintankProbeHTTPS) Run() (CheckResult, error) {
	deadline := time.Now().Add(p.Timeout)
	result := &HTTPSResult{}

	// reader
	tmpPort := ""
	if p.Port != 443 {
		tmpPort = fmt.Sprintf(":%d", p.Port)
	}
	url := fmt.Sprintf("https://%s%s%s", p.Host, tmpPort, p.Path)
	sendBody := bytes.NewReader([]byte(p.Body))
	request, err := http.NewRequest(p.Method, url, sendBody)

	if err != nil {
		msg := fmt.Sprintf("Invalid request settings. %s", err.Error())
		result.Error = &msg
		return result, nil
	}

	// Parsing header (use fake request)
	if p.Headers != "" {
		headReader := bufio.NewReader(strings.NewReader("GET / HTTP/1.1\r\n" + p.Headers + "\r\n\r\n"))
		dummyRequest, err := http.ReadRequest(headReader)
		if err != nil {
			msg := err.Error()
			result.Error = &msg
			return result, nil
		}

		for key := range dummyRequest.Header {
			request.Header.Set(key, dummyRequest.Header.Get(key))
		}

		if dummyRequest.Host != "" {
			request.Host = dummyRequest.Host
		}
	}

	if _, found := request.Header["Accept-Encoding"]; !found {
		request.Header.Set("Accept-Encoding", "gzip")
	}

	// always close the connection
	request.Header.Set("Connection", "close")

	// DNS lookup
	step := time.Now()
	addrs, err := net.LookupHost(p.Host)
	if err != nil || len(addrs) < 1 {
		msg := "failed to resolve hostname to IP."
		result.Error = &msg
		return result, nil
	}

	dnsResolve := time.Since(step).Seconds() * 1000
	result.DNS = &dnsResolve

	tlsConfig := &tls.Config{
		InsecureSkipVerify: !p.ValidateCert,
		ServerName:         p.Host,
	}

	// Dialing
	start := time.Now()
	tcpconn, err := net.DialTimeout("tcp", net.JoinHostPort(addrs[0], strconv.FormatInt(p.Port, 10)), p.Timeout)
	if err != nil {
		opError, ok := err.(*net.OpError)
		if ok {
			if opError.Timeout() {
				msg := "timeout while connecting to host."
				result.Error = &msg
				return result, nil
			}
			msg := fmt.Sprintf("%s error. %s", opError.Op, opError.Err.Error())
			result.Error = &msg
			return result, nil
		}
		msg := err.Error()
		result.Error = &msg
		return result, nil
	}
	tcpconn.SetDeadline(deadline)

	conn := tls.Client(tcpconn, tlsConfig)
	err = conn.Handshake()
	if err != nil {
		msg := err.Error()
		result.Error = &msg
		return result, nil
	}
	// Read certificate
	certs := conn.ConnectionState().PeerCertificates

	defer conn.Close()
	connecting := time.Since(start).Seconds() * 1000
	result.Connect = &connecting

	// Send
	step = time.Now()
	if err := request.Write(conn); err != nil {
		opError, ok := err.(*net.OpError)
		if ok {
			if opError.Timeout() {
				msg := "timeout while sending request."
				result.Error = &msg
				return result, nil
			}
			msg := fmt.Sprintf("%s error. %s", opError.Op, opError.Err.Error())
			result.Error = &msg
			return result, nil
		}
		msg := err.Error()
		result.Error = &msg
		return result, nil
	}
	send := time.Since(step).Seconds() * 1000
	result.Send = &send

	// Wait.
	// ReadResponse will block until the headers are received,
	// it will then return our response struct without wiating for the entire
	// response to be read from the connection.
	step = time.Now()
	response, err := http.ReadResponse(bufio.NewReader(conn), request)
	if err != nil {
		opError, ok := err.(*net.OpError)
		if ok {
			if opError.Timeout() {
				msg := "timeout while waiting for response."
				result.Error = &msg
				return result, nil
			}
			msg := fmt.Sprintf("%s error. %s", opError.Op, opError.Err.Error())
			result.Error = &msg
			return result, nil
		}
		msg := err.Error()
		result.Error = &msg
		return result, nil
	}
	wait := time.Since(step).Seconds() * 1000
	result.Wait = &wait

	step = time.Now()
	var body bytes.Buffer
	data := make([]byte, 1024)
	for {
		n, err := response.Body.Read(data)
		body.Write(data[:n])
		if err != nil {
			if err != io.EOF {
				opError, ok := err.(*net.OpError)
				if ok {
					if opError.Timeout() {
						msg := "timeout while receiving response."
						result.Error = &msg
						return result, nil
					}
					msg := fmt.Sprintf("%s error. %s", opError.Op, opError.Err.Error())
					result.Error = &msg
					return result, nil
				}
				msg := err.Error()
				result.Error = &msg
				return result, nil
			} else {
				break
			}
		}

		if int64(body.Len()) > p.DownloadLimit {
			conn.Close()
			break
		}
	}

	recv := time.Since(step).Seconds() * 1000
	/*
		Total time
	*/
	total := time.Since(start).Seconds() * 1000
	result.Total = &total

	result.Recv = &recv

	// Data Length
	var headers bytes.Buffer
	response.Header.Write(&headers)
	dataLength := float64(body.Len() + headers.Len())
	result.DataLength = &dataLength

	//throughput
	if recv > 0 {
		throughput := float64(body.Len()) / (recv / 1000.0)
		result.Throughput = &throughput
	}

	if err != nil {
		msg := err.Error()
		result.Error = &msg
		return result, nil
	}

	// Error response
	statusCode := float64(response.StatusCode)
	result.StatusCode = &statusCode
	if statusCode >= 400 {
		msg := "Invalid status code " + strconv.Itoa(response.StatusCode)
		result.Error = &msg
		return result, nil
	}

	if certs == nil || len(certs) < 1 {
		log.Printf("no PeerCertificates for connection to %s", p.Host)
	} else {
		timeTilExpiry := certs[0].NotAfter.Sub(time.Now())
		secondsTilExpiry := float64(timeTilExpiry) / float64(time.Second)
		result.Expiry = &secondsTilExpiry
	}

	// Regex
	if p.ExpectRegex != "" {
		rgx, err := regexp.Compile(p.ExpectRegex)
		if err != nil {
			msg := err.Error()
			result.Error = &msg
			return result, nil
		}

		// Handle gzip
		var decodedBody string
		switch response.Header.Get("Content-Encoding") {
		case "gzip":
			reader, err := gzip.NewReader(&body)
			if err != nil {
				msg := err.Error()
				result.Error = &msg
				return result, nil
			}
			decodedBodyBytes, err := ioutil.ReadAll(reader)
			if err != nil && len(decodedBodyBytes) == 0 {
				msg := err.Error()
				result.Error = &msg
				return result, nil
			}
			decodedBody = string(decodedBodyBytes)
		default:
			decodedBody = body.String()
		}

		if !rgx.MatchString(decodedBody) {
			msg := "expectRegex did not match"
			result.Error = &msg
			return result, nil
		}
	}

	return result, nil
}
