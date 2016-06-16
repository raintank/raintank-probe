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

	"github.com/raintank/raintank-metric/schema"
	"github.com/raintank/raintank-probe/probe"
	m "github.com/raintank/worldping-api/pkg/models"
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

func (r *HTTPSResult) Metrics(t time.Time, check *m.MonitorDTO) []*schema.MetricData {
	metrics := make([]*schema.MetricData, 0)
	if r.DNS != nil {
		metrics = append(metrics, &schema.MetricData{
			OrgId:      int(check.OrgId),
			Name:       fmt.Sprintf("worldping.%s.%s.https.dns", check.EndpointSlug, probe.Self.Slug),
			Metric:     "worldping.https.dns",
			Interval:   int(check.Frequency),
			Unit:       "ms",
			TargetType: "gauge",
			Time:       t.Unix(),
			Tags: []string{
				fmt.Sprintf("endpoint:%s", check.EndpointSlug),
				fmt.Sprintf("probe:%s", probe.Self.Slug),
				"checkType:https",
			},
			Value: *r.DNS,
		})
	}
	if r.Connect != nil {
		metrics = append(metrics, &schema.MetricData{
			OrgId:      int(check.OrgId),
			Name:       fmt.Sprintf("worldping.%s.%s.https.connect", check.EndpointSlug, probe.Self.Slug),
			Metric:     "worldping.https.connect",
			Interval:   int(check.Frequency),
			Unit:       "ms",
			TargetType: "gauge",
			Time:       t.Unix(),
			Tags: []string{
				fmt.Sprintf("endpoint:%s", check.EndpointSlug),
				fmt.Sprintf("probe:%s", probe.Self.Slug),
				"checkType:https",
			},
			Value: *r.Connect,
		})
	}
	if r.Send != nil {
		metrics = append(metrics, &schema.MetricData{
			OrgId:      int(check.OrgId),
			Name:       fmt.Sprintf("worldping.%s.%s.https.send", check.EndpointSlug, probe.Self.Slug),
			Metric:     "worldping.https.send",
			Interval:   int(check.Frequency),
			Unit:       "ms",
			TargetType: "gauge",
			Time:       t.Unix(),
			Tags: []string{
				fmt.Sprintf("endpoint:%s", check.EndpointSlug),
				fmt.Sprintf("probe:%s", probe.Self.Slug),
				"checkType:https",
			},
			Value: *r.Send,
		})
	}
	if r.Wait != nil {
		metrics = append(metrics, &schema.MetricData{
			OrgId:      int(check.OrgId),
			Name:       fmt.Sprintf("worldping.%s.%s.https.wait", check.EndpointSlug, probe.Self.Slug),
			Metric:     "worldping.https.wait",
			Interval:   int(check.Frequency),
			Unit:       "ms",
			TargetType: "gauge",
			Time:       t.Unix(),
			Tags: []string{
				fmt.Sprintf("endpoint:%s", check.EndpointSlug),
				fmt.Sprintf("probe:%s", probe.Self.Slug),
				"checkType:https",
			},
			Value: *r.Wait,
		})
	}
	if r.Recv != nil {
		metrics = append(metrics, &schema.MetricData{
			OrgId:      int(check.OrgId),
			Name:       fmt.Sprintf("worldping.%s.%s.https.recv", check.EndpointSlug, probe.Self.Slug),
			Metric:     "worldping.https.recv",
			Interval:   int(check.Frequency),
			Unit:       "ms",
			TargetType: "gauge",
			Time:       t.Unix(),
			Tags: []string{
				fmt.Sprintf("endpoint:%s", check.EndpointSlug),
				fmt.Sprintf("probe:%s", probe.Self.Slug),
				"checkType:https",
			},
			Value: *r.Recv,
		})
	}
	if r.Total != nil {
		metrics = append(metrics, &schema.MetricData{
			OrgId:      int(check.OrgId),
			Name:       fmt.Sprintf("worldping.%s.%s.https.total", check.EndpointSlug, probe.Self.Slug),
			Metric:     "worldping.https.total",
			Interval:   int(check.Frequency),
			Unit:       "ms",
			TargetType: "gauge",
			Time:       t.Unix(),
			Tags: []string{
				fmt.Sprintf("endpoint:%s", check.EndpointSlug),
				fmt.Sprintf("probe:%s", probe.Self.Slug),
				"checkType:https",
			},
			Value: *r.Total,
		})
		metrics = append(metrics, &schema.MetricData{
			OrgId:      int(check.OrgId),
			Name:       fmt.Sprintf("worldping.%s.%s.https.default", check.EndpointSlug, probe.Self.Slug),
			Metric:     "worldping.https.default",
			Interval:   int(check.Frequency),
			Unit:       "ms",
			TargetType: "gauge",
			Time:       t.Unix(),
			Tags: []string{
				fmt.Sprintf("endpoint:%s", check.EndpointSlug),
				fmt.Sprintf("probe:%s", probe.Self.Slug),
				"checkType:https",
			},
			Value: *r.Total,
		})
	}
	if r.Throughput != nil {
		metrics = append(metrics, &schema.MetricData{
			OrgId:      int(check.OrgId),
			Name:       fmt.Sprintf("worldping.%s.%s.https.throughput", check.EndpointSlug, probe.Self.Slug),
			Metric:     "worldping.https.throughput",
			Interval:   int(check.Frequency),
			Unit:       "bytes",
			TargetType: "rate",
			Time:       t.Unix(),
			Tags: []string{
				fmt.Sprintf("endpoint:%s", check.EndpointSlug),
				fmt.Sprintf("probe:%s", probe.Self.Slug),
				"checkType:https",
			},
			Value: *r.Throughput,
		})
	}
	if r.DataLength != nil {
		metrics = append(metrics, &schema.MetricData{
			OrgId:      int(check.OrgId),
			Name:       fmt.Sprintf("worldping.%s.%s.https.dataLength", check.EndpointSlug, probe.Self.Slug),
			Metric:     "worldping.https.dataLength",
			Interval:   int(check.Frequency),
			Unit:       "bytess",
			TargetType: "gauge",
			Time:       t.Unix(),
			Tags: []string{
				fmt.Sprintf("endpoint:%s", check.EndpointSlug),
				fmt.Sprintf("probe:%s", probe.Self.Slug),
				"checkType:https",
			},
			Value: *r.DataLength,
		})
	}
	if r.StatusCode != nil {
		metrics = append(metrics, &schema.MetricData{
			OrgId:      int(check.OrgId),
			Name:       fmt.Sprintf("worldping.%s.%s.https.statusCode", check.EndpointSlug, probe.Self.Slug),
			Metric:     "worldping.https.statusCode",
			Interval:   int(check.Frequency),
			Unit:       "code",
			TargetType: "gauge",
			Time:       t.Unix(),
			Tags: []string{
				fmt.Sprintf("endpoint:%s", check.EndpointSlug),
				fmt.Sprintf("probe:%s", probe.Self.Slug),
				"checkType:https",
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
	if !ok {
		return nil, fmt.Errorf("no timeout passed.")
	}
	t, ok := timeout.(float64)
	if !ok {
		return nil, fmt.Errorf("invalid value for timeout, must be number.")
	}
	if t <= 0.0 {
		return nil, fmt.Errorf("invalid value for timeout, must be greater then 0.")
	}
	p.Timeout = time.Duration(time.Millisecond * time.Duration(int(1000.0*t)))

	port, ok := settings["port"]
	if !ok {
		p.Port = 443
	} else {
		p.Port, ok = port.(int64)
		if !ok {
			return nil, fmt.Errorf("invalid value for port, must be integer.")
		}
	}
	if p.Port < 1 || p.Port > 65535 {
		return nil, fmt.Errorf("invalid port number.")
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

	tlsConfig := &tls.Config{
		InsecureSkipVerify: !p.ValidateCert,
		ServerName:         p.Host,
	}

	// Dialing
	start := time.Now()
	tcpconn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", addrs[0], p.Port), p.Timeout)
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

	// Wait
	step = time.Now()

	//read first byte data
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

		if buf.Len() >= limit {
			conn.Close()
			break
		}
	}

	recv := time.Since(step).Seconds() * 1000
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
				log.Printf("failed to decode content body for request to %s. %s", url, err)
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

	if certs == nil || len(certs) < 1 {
		log.Printf("no PeerCerticates for connection to %s", p.Host)
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

		if !rgx.MatchString(string(body)) {
			msg := "expectRegex did not match"
			result.Error = &msg
			return result, nil
		}
	}

	return result, nil
}
