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
	m "github.com/raintank/worldping-api/pkg/models"
	"gopkg.in/raintank/schema.v1"
)

// HTTPResult struct
type HTTPResult struct {
	DNS        *float64
	Connect    *float64
	Send       *float64
	Wait       *float64
	Recv       *float64
	Total      *float64
	DataLength *float64
	Throughput *float64
	StatusCode *float64
	Error      *string
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
				fmt.Sprintf("endpoint_id:%d", check.EndpointId),
				fmt.Sprintf("monitor_id:%d", check.Id),
				fmt.Sprintf("collector:%s", probe.Self.Slug),
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
				fmt.Sprintf("endpoint_id:%d", check.EndpointId),
				fmt.Sprintf("monitor_id:%d", check.Id),
				fmt.Sprintf("collector:%s", probe.Self.Slug),
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
				fmt.Sprintf("endpoint_id:%d", check.EndpointId),
				fmt.Sprintf("monitor_id:%d", check.Id),
				fmt.Sprintf("collector:%s", probe.Self.Slug),
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
				fmt.Sprintf("endpoint_id:%d", check.EndpointId),
				fmt.Sprintf("monitor_id:%d", check.Id),
				fmt.Sprintf("collector:%s", probe.Self.Slug),
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
				fmt.Sprintf("endpoint_id:%d", check.EndpointId),
				fmt.Sprintf("monitor_id:%d", check.Id),
				fmt.Sprintf("collector:%s", probe.Self.Slug),
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
				fmt.Sprintf("endpoint_id:%d", check.EndpointId),
				fmt.Sprintf("monitor_id:%d", check.Id),
				fmt.Sprintf("collector:%s", probe.Self.Slug),
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
				fmt.Sprintf("endpoint_id:%d", check.EndpointId),
				fmt.Sprintf("monitor_id:%d", check.Id),
				fmt.Sprintf("collector:%s", probe.Self.Slug),
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
				fmt.Sprintf("endpoint_id:%d", check.EndpointId),
				fmt.Sprintf("monitor_id:%d", check.Id),
				fmt.Sprintf("collector:%s", probe.Self.Slug),
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
				fmt.Sprintf("endpoint_id:%d", check.EndpointId),
				fmt.Sprintf("monitor_id:%d", check.Id),
				fmt.Sprintf("collector:%s", probe.Self.Slug),
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
				fmt.Sprintf("endpoint_id:%d", check.EndpointId),
				fmt.Sprintf("monitor_id:%d", check.Id),
				fmt.Sprintf("collector:%s", probe.Self.Slug),
			},
			Value: *r.StatusCode,
		})
	}
	return metrics
}

// RaintankProbeHTTP struct.
type RaintankProbeHTTP struct {
	Host        string
	Path        string
	Port        int64
	Method      string
	Headers     string
	ExpectRegex string
	Body        string
	Timeout     time.Duration
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

	return &p, nil
}

// Run checking
func (p *RaintankProbeHTTP) Run() (CheckResult, error) {
	deadline := time.Now().Add(time.Second * time.Duration(p.Timeout))
	result := &HTTPResult{}

	// reader
	tmpPort := ""
	if p.Port != 80 {
		tmpPort = fmt.Sprintf(":%d", p.Port)
	}
	url := fmt.Sprintf("http://%s%s%s", p.Host, tmpPort, p.Path)
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
		result.Error = &msg
		return result, nil
	}
	dnsResolve := time.Since(step).Seconds() * 1000
	result.DNS = &dnsResolve

	// Dialing
	start := time.Now()
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", addrs[0], p.Port), p.Timeout)
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
	conn.SetDeadline(deadline)
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

	// Wait
	step = time.Now()

	// read first byte data
	firstData := make([]byte, 1)
	_, err = conn.Read(firstData)
	if err != nil {
		if err != io.EOF {
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
	}
	wait := time.Since(step).Seconds() * 1000
	result.Wait = &wait

	// Receive
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

		if buf.Len() > limit {
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
	dataLength := float64(buf.Len())
	result.DataLength = &dataLength

	//throughput
	if recv > 0 && dataLength > 0 {
		throughput := dataLength / (recv / 1000.0)
		result.Throughput = &throughput
	}

	response, err := http.ReadResponse(bufio.NewReader(&buf), request)

	if err != nil {
		msg := err.Error()
		result.Error = &msg
		return result, nil
	}

	var reader io.ReadCloser
	if p.ExpectRegex != "" {
		// Handle gzip
		switch response.Header.Get("Content-Encoding") {
		case "gzip":
			reader, err = gzip.NewReader(response.Body)
			if err != nil {
				msg := err.Error()
				result.Error = &msg
				return result, nil
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

	// Regex
	if p.ExpectRegex != "" {
		rgx, err := regexp.Compile(p.ExpectRegex)
		if err != nil {
			msg := err.Error()
			result.Error = &msg
			return result, nil
		}

		if !rgx.MatchString(string(body)) {
			msg := "expectRegex did not match"
			result.Error = &msg
			return result, nil
		}
	}

	return result, nil
}
