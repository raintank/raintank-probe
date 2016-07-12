package checks

import (
	"time"

	"github.com/raintank/raintank-metric/schema"
	m "github.com/raintank/worldping-api/pkg/models"
)

type CheckResult interface {
	Metrics(time.Time, *m.CheckWithSlug) []*schema.MetricData
	ErrorMsg() string
}
