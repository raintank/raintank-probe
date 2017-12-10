package models

import (
	"errors"
	"github.com/raintank/worldping-api/pkg/setting"
	"time"
)

var ErrInvalidQuotaTarget = errors.New("Invalid quota target")

type Quota struct {
	Id      int64
	OrgId   int64
	Target  string
	Limit   int64
	Created time.Time
	Updated time.Time
}

type QuotaScope struct {
	Name         string
	Target       string
	DefaultLimit int64
}

type OrgQuotaDTO struct {
	OrgId  int64  `json:"org_id"`
	Target string `json:"target"`
	Limit  int64  `json:"limit"`
	Used   int64  `json:"used"`
}

type GlobalQuotaDTO struct {
	Target string `json:"target"`
	Limit  int64  `json:"limit"`
	Used   int64  `json:"used"`
}

func GetQuotaScopes(target string) ([]QuotaScope, error) {
	scopes := make([]QuotaScope, 0)
	switch target {
	case "endpoint":
		scopes = append(scopes,
			QuotaScope{Name: "global", Target: target, DefaultLimit: setting.Quota.Global.Endpoint},
			QuotaScope{Name: "org", Target: target, DefaultLimit: setting.Quota.Org.Endpoint},
		)
		return scopes, nil
	case "probe":
		scopes = append(scopes,
			QuotaScope{Name: "global", Target: target, DefaultLimit: setting.Quota.Global.Probe},
			QuotaScope{Name: "org", Target: target, DefaultLimit: setting.Quota.Org.Probe},
		)
		return scopes, nil
	case "downloadLimit":
		scopes = append(scopes,
			QuotaScope{Name: "org", Target: target, DefaultLimit: setting.Quota.Org.DownloadLimit},
		)
		return scopes, nil
	default:
		return scopes, ErrInvalidQuotaTarget
	}
}
