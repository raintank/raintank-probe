package models

import (
	"fmt"
	"time"
)

// Job is a job for an alert execution
// note that LastPointTs is a time denoting the timestamp of the last point to run against
// this way the check runs always on the right data, irrespective of execution delays
// that said, for convenience, we track the generatedAt timestamp
type AlertingJob struct {
	*CheckForAlertDTO
	GeneratedAt time.Time
	LastPointTs time.Time
	NewState    CheckEvalResult
	TimeExec    time.Time
}

func (job *AlertingJob) String() string {
	return fmt.Sprintf("<Job> checkId=%d generatedAt=%s lastPointTs=%s definition: %d probes for %d steps", job.Id, job.GeneratedAt, job.LastPointTs, job.HealthSettings.NumProbes, job.HealthSettings.Steps)
}
