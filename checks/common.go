package checks

import (
	"time"

	"github.com/raintank/raintank-metric/schema"
	m "github.com/raintank/worldping-api/pkg/models"
)

type CheckResult interface {
	Metrics(time.Time, *m.MonitorDTO) []*schema.MetricData
	ErrorMsg() string
}
