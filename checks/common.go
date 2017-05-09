package checks

import (
	"fmt"
	"net"
	"strings"
	"time"

	m "github.com/raintank/worldping-api/pkg/models"
	"gopkg.in/raintank/schema.v1"
)

type CheckResult interface {
	Metrics(time.Time, *m.CheckWithSlug) []*schema.MetricData
	ErrorMsg() string
}

func ResolveHost(host, ipversion string) (string, error) {
	addrs, err := net.LookupHost(host)
	if err != nil || len(addrs) < 1 {
		return "", fmt.Errorf("failed to resolve hostname to IP.")
	}

	for _, addr := range addrs {
		if ipversion == "any" {
			return addr, nil
		}

		if strings.Contains(addr, ":") || strings.Contains(addr, "%") {
			if ipversion == "v6" {
				return addr, nil
			}
		} else {
			if ipversion == "v4" {
				return addr, nil
			}
		}
	}

	return "", fmt.Errorf("failed to resolve hostname to valid IP.")
}
