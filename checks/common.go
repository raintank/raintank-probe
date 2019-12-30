package checks

import (
	"fmt"
	"net"
	"time"

	m "github.com/raintank/worldping-api/pkg/models"
	"github.com/grafana/metrictank/schema"
)

type CheckResult interface {
	Metrics(time.Time, *m.CheckWithSlug) []*schema.MetricData
	ErrorMsg() string
}

func ResolveHost(host, ipversion string) (string, error) {
	addrs, err := net.LookupIP(host)
	if err != nil || len(addrs) < 1 {
		return "", fmt.Errorf("failed to resolve hostname to IP.")
	}

	for _, addr := range addrs {
		// only allow Global unicast, or loopback addresses
		// to be used.
		if !(addr.IsGlobalUnicast() || addr.IsLoopback()) {
			continue
		}
		if ipversion == "any" {
			return addr.String(), nil
		}

		if !isIPv4(addr) {
			if ipversion == "v6" {
				return addr.String(), nil
			}
		} else {
			if ipversion == "v4" {
				return addr.String(), nil
			}
		}
	}

	return "", fmt.Errorf("failed to resolve hostname to valid IP.")
}

func isIPv4(ip net.IP) bool {
	ip4 := ip.To4()
	return ip4 != nil
}
