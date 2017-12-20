package checks

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/raintank/raintank-probe/probe"
	"github.com/raintank/worldping-api/pkg/log"
	m "github.com/raintank/worldping-api/pkg/models"
	"gopkg.in/raintank/schema.v1"
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
	Throughput *float64 `json:"throughput"`
	StatusCode *float64 `json:"statusCode"`
	Error      *string  `json:"error"`
}

func (r *HTTPResult) ErrorMsg() string {
	if r.Error == nil {
		return ""
	}
	return *r.Error
}

func (r *HTTPResult) Metrics(t time.Time, check *m.CheckWithSlug) []*schema.MetricData {
	metrics := make([]*schema.MetricData, 0)
	if r.DNS != nil {
		metrics = append(metrics, &schema.MetricData{
			OrgId:    int(check.OrgId),
			Name:     fmt.Sprintf("worldping.%s.%s.http.dns", check.Slug, probe.Self.Slug),
			Metric:   "worldping.http.dns",
			Interval: int(check.Frequency),
			Unit:     "ms",
			Mtype:    "gauge",
			Time:     t.Unix(),
			Tags: []string{
				fmt.Sprintf("endpoint=%s", check.Slug),
				fmt.Sprintf("monitor_type=%s", check.Type),
				fmt.Sprintf("probe=%s", probe.Self.Slug),
			},
			Value: *r.DNS,
		})
	}
	if r.Connect != nil {
		metrics = append(metrics, &schema.MetricData{
			OrgId:    int(check.OrgId),
			Name:     fmt.Sprintf("worldping.%s.%s.http.connect", check.Slug, probe.Self.Slug),
			Metric:   "worldping.http.connect",
			Interval: int(check.Frequency),
			Unit:     "ms",
			Mtype:    "gauge",
			Time:     t.Unix(),
			Tags: []string{
				fmt.Sprintf("endpoint=%s", check.Slug),
				fmt.Sprintf("monitor_type=%s", check.Type),
				fmt.Sprintf("probe=%s", probe.Self.Slug),
			},
			Value: *r.Connect,
		})
	}
	if r.Send != nil {
		metrics = append(metrics, &schema.MetricData{
			OrgId:    int(check.OrgId),
			Name:     fmt.Sprintf("worldping.%s.%s.http.send", check.Slug, probe.Self.Slug),
			Metric:   "worldping.http.send",
			Interval: int(check.Frequency),
			Unit:     "ms",
			Mtype:    "gauge",
			Time:     t.Unix(),
			Tags: []string{
				fmt.Sprintf("endpoint=%s", check.Slug),
				fmt.Sprintf("monitor_type=%s", check.Type),
				fmt.Sprintf("probe=%s", probe.Self.Slug),
			},
			Value: *r.Send,
		})
	}
	if r.Wait != nil {
		metrics = append(metrics, &schema.MetricData{
			OrgId:    int(check.OrgId),
			Name:     fmt.Sprintf("worldping.%s.%s.http.wait", check.Slug, probe.Self.Slug),
			Metric:   "worldping.http.wait",
			Interval: int(check.Frequency),
			Unit:     "ms",
			Mtype:    "gauge",
			Time:     t.Unix(),
			Tags: []string{
				fmt.Sprintf("endpoint=%s", check.Slug),
				fmt.Sprintf("monitor_type=%s", check.Type),
				fmt.Sprintf("probe=%s", probe.Self.Slug),
			},
			Value: *r.Wait,
		})
	}
	if r.Recv != nil {
		metrics = append(metrics, &schema.MetricData{
			OrgId:    int(check.OrgId),
			Name:     fmt.Sprintf("worldping.%s.%s.http.recv", check.Slug, probe.Self.Slug),
			Metric:   "worldping.http.recv",
			Interval: int(check.Frequency),
			Unit:     "ms",
			Mtype:    "gauge",
			Time:     t.Unix(),
			Tags: []string{
				fmt.Sprintf("endpoint=%s", check.Slug),
				fmt.Sprintf("monitor_type=%s", check.Type),
				fmt.Sprintf("probe=%s", probe.Self.Slug),
			},
			Value: *r.Recv,
		})
	}
	if r.Total != nil {
		metrics = append(metrics, &schema.MetricData{
			OrgId:    int(check.OrgId),
			Name:     fmt.Sprintf("worldping.%s.%s.http.total", check.Slug, probe.Self.Slug),
			Metric:   "worldping.http.total",
			Interval: int(check.Frequency),
			Unit:     "ms",
			Mtype:    "gauge",
			Time:     t.Unix(),
			Tags: []string{
				fmt.Sprintf("endpoint=%s", check.Slug),
				fmt.Sprintf("monitor_type=%s", check.Type),
				fmt.Sprintf("probe=%s", probe.Self.Slug),
			},
			Value: *r.Total,
		})
		metrics = append(metrics, &schema.MetricData{
			OrgId:    int(check.OrgId),
			Name:     fmt.Sprintf("worldping.%s.%s.http.default", check.Slug, probe.Self.Slug),
			Metric:   "worldping.http.default",
			Interval: int(check.Frequency),
			Unit:     "ms",
			Mtype:    "gauge",
			Time:     t.Unix(),
			Tags: []string{
				fmt.Sprintf("endpoint=%s", check.Slug),
				fmt.Sprintf("monitor_type=%s", check.Type),
				fmt.Sprintf("probe=%s", probe.Self.Slug),
			},
			Value: *r.Total,
		})
	}
	if r.Throughput != nil {
		metrics = append(metrics, &schema.MetricData{
			OrgId:    int(check.OrgId),
			Name:     fmt.Sprintf("worldping.%s.%s.http.throughput", check.Slug, probe.Self.Slug),
			Metric:   "worldping.http.throughput",
			Interval: int(check.Frequency),
			Unit:     "B/s",
			Mtype:    "rate",
			Time:     t.Unix(),
			Tags: []string{
				fmt.Sprintf("endpoint=%s", check.Slug),
				fmt.Sprintf("monitor_type=%s", check.Type),
				fmt.Sprintf("probe=%s", probe.Self.Slug),
			},
			Value: *r.Throughput,
		})
	}
	if r.DataLength != nil {
		metrics = append(metrics, &schema.MetricData{
			OrgId:    int(check.OrgId),
			Name:     fmt.Sprintf("worldping.%s.%s.http.dataLength", check.Slug, probe.Self.Slug),
			Metric:   "worldping.http.dataLength",
			Interval: int(check.Frequency),
			Unit:     "B",
			Mtype:    "gauge",
			Time:     t.Unix(),
			Tags: []string{
				fmt.Sprintf("endpoint=%s", check.Slug),
				fmt.Sprintf("monitor_type=%s", check.Type),
				fmt.Sprintf("probe=%s", probe.Self.Slug),
			},
			Value: *r.DataLength,
		})
	}
	if r.StatusCode != nil {
		metrics = append(metrics, &schema.MetricData{
			OrgId:    int(check.OrgId),
			Name:     fmt.Sprintf("worldping.%s.%s.http.statusCode", check.Slug, probe.Self.Slug),
			Metric:   "worldping.http.statusCode",
			Interval: int(check.Frequency),
			Unit:     "",
			Mtype:    "gauge",
			Time:     t.Unix(),
			Tags: []string{
				fmt.Sprintf("endpoint=%s", check.Slug),
				fmt.Sprintf("monitor_type=%s", check.Type),
				fmt.Sprintf("probe=%s", probe.Self.Slug),
			},
			Value: *r.StatusCode,
		})
	}
	return metrics
}

// RaintankProbeHTTP struct.
type RaintankProbeHTTP struct {
	Host          string        `json:"host"`
	Path          string        `json:"path"`
	Port          int64         `json:"port"`
	Method        string        `json:"method"`
	Headers       string        `json:"headers"`
	ExpectRegex   string        `json:"expectRegex"`
	Body          string        `json:"body"`
	Timeout       time.Duration `json:"timeout"`
	DownloadLimit int64         `json:"downloadLimit"`
	IPVersion     string        `json:"ipversion"`
}

// NewRaintankHTTPProbe json check
func NewRaintankHTTPProbe(settings map[string]interface{}) (*RaintankProbeHTTP, error) {
	p := RaintankProbeHTTP{}
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
		p.Port = 80
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

	version, ok := settings["ipversion"]
	if !ok {
		p.IPVersion = "v4"
	} else {
		p.IPVersion, ok = version.(string)
		if !ok {
			return nil, fmt.Errorf("invalid value for ipversion, must be string.")
		}
	}
	if !(p.IPVersion == "v4" || p.IPVersion == "v6" || p.IPVersion == "any") {
		return nil, fmt.Errorf("ipversion must be v4, v6, or any.")
	}

	return &p, nil
}

// Run checking
func (p *RaintankProbeHTTP) Run() (CheckResult, error) {
	deadline := time.Now().Add(p.Timeout)
	result := &HTTPResult{}

	// reader
	tmpHost := p.Host
	if p.Port != 80 {
		tmpHost = net.JoinHostPort(p.Host, strconv.FormatInt(p.Port, 10))
	} else if strings.Contains(p.Host, ":") || strings.Contains(p.Host, "%") {
		tmpHost = "[" + p.Host + "]"
	}
	url := fmt.Sprintf("http://%s%s", tmpHost, p.Path)
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

	ipAddr, err := ResolveHost(p.Host, p.IPVersion)
	if err != nil {
		msg := fmt.Sprintf("error resolving hostname. %s", err.Error())
		result.Error = &msg
		return result, nil
	}
	if time.Now().After(deadline) {
		msg := "error resolving hostname. timeout"
		result.Error = &msg
		return result, nil
	}

	dnsResolve := time.Since(step).Seconds() * 1000
	result.DNS = &dnsResolve

	// fmt.Printf("http connecting to %#v\n", ipAddr)

	// Dialing
	start := time.Now()
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(ipAddr, strconv.FormatInt(p.Port, 10)), p.Timeout)
	if err != nil {
		msg := ""
		opError, ok := err.(*net.OpError)
		if ok {
			if opError.Timeout() {
				msg = "error connecting. timeout"
			} else {
				msg = fmt.Sprintf("error connecting. . %s", err.Error())
			}
		} else {
			msg = fmt.Sprintf("error connecting. . %s", err.Error())
		}

		result.Error = &msg
		return result, nil
	}

	conn.SetDeadline(deadline)
	defer conn.Close()

	connecting := time.Since(start).Seconds() * 1000
	result.Connect = &connecting

	// Send
	step = time.Now()
	if err := request.Write(conn); err != nil {
		msg := ""
		opError, ok := err.(*net.OpError)
		if ok {
			if opError.Timeout() {
				msg = "error sending request. timeout"
			} else {
				msg = fmt.Sprintf("error sending request. %s", err.Error())
			}
		} else {
			msg = fmt.Sprintf("error sending request. %s", err.Error())
		}

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
		msg := ""
		opError, ok := err.(*net.OpError)
		if ok {
			if opError.Timeout() {
				msg = "error reading response. timeout"
			} else {
				msg = fmt.Sprintf("error reading response. %s", err.Error())
			}
		} else {
			msg = fmt.Sprintf("error reading response. %s", err.Error())
		}

		result.Error = &msg
		return result, nil
	}

	defer response.Body.Close()

	wait := time.Since(step).Seconds() * 1000
	result.Wait = &wait

	step = time.Now()
	var body bytes.Buffer
	data := make([]byte, 1024)
	for {
		n, err := response.Body.Read(data)
		body.Write(data[:n])
		if err == io.EOF {
			break
		}

		if err != nil {
			msg := ""

			opError, ok := err.(*net.OpError)
			if ok {
				if opError.Timeout() {
					msg = "error reading response. timeout"
				} else {
					msg = fmt.Sprintf("error reading response. %s", err.Error())
				}
			} else {
				msg = fmt.Sprintf("error reading response. %s", err.Error())
			}

			result.Error = &msg
			return result, nil
		}

		if int64(body.Len()) > p.DownloadLimit {
			break
		}
	}

	recv := time.Since(step).Seconds() * 1000
	result.Recv = &recv

	response.Body.Close()
	conn.Close()

	/*
		Total time
	*/
	total := time.Since(start).Seconds() * 1000
	result.Total = &total

	// Data Length
	var headers bytes.Buffer
	response.Header.Write(&headers)
	dataLength := float64(body.Len() + headers.Len())
	result.DataLength = &dataLength

	// throughput
	if recv > 0 {
		throughput := float64(body.Len()) / (recv / 1000.0)
		result.Throughput = &throughput
	}

	// Error response
	statusCode := float64(response.StatusCode)
	result.StatusCode = &statusCode
	if statusCode >= 400 {
		msg := "Invalid status code " + strconv.Itoa(response.StatusCode)
		result.Error = &msg
		return result, nil
	}

	// Regex
	if p.ExpectRegex != "" {
		rgx, err := regexp.Compile(p.ExpectRegex)
		if err != nil {
			msg := fmt.Sprintf("expectRegex error. %s", err.Error())

			result.Error = &msg
			return result, nil
		}

		// Handle gzip
		var decodedBody string

		switch strings.ToLower(response.Header.Get("Content-Encoding")) {
		case "gzip":
			reader, err := gzip.NewReader(&body)
			if err != nil {
				msg := fmt.Sprintf("error decoding content. %s", err.Error())

				result.Error = &msg
				return result, nil
			}

			decodedBodyBytes, err := ioutil.ReadAll(reader)
			if err != nil && len(decodedBodyBytes) == 0 {
				msg := fmt.Sprintf("error decoding content. %s", err.Error())

				result.Error = &msg
				return result, nil
			}
			decodedBody = string(decodedBodyBytes)
		case "", "identity":
			decodedBody = body.String()
		default:
			msg := "unrecognized Content-Encoding: " + response.Header.Get("Content-Encoding")
			result.Error = &msg
			return result, nil
		}

		if !rgx.MatchString(decodedBody) {
			log.Debug("expectRegex %s did not match returned body %s", p.ExpectRegex, decodedBody)

			msg := "expectRegex did not match"
			result.Error = &msg
			return result, nil
		}
	}

	return result, nil
}
