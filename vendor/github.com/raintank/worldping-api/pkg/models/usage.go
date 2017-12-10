package models

type Usage struct {
	Endpoints EndpointUsage
	Probes    ProbeUsage
	Checks    CheckUsage
}

type EndpointUsage struct {
	Total  int64
	PerOrg map[string]int64
}

type ProbeUsage struct {
	Total  int64
	PerOrg map[string]int64
}

type CheckUsage struct {
	Total int64
	HTTP  CheckHTTPUsage
	HTTPS CheckHTTPSUsage
	PING  CheckPINGUsage
	DNS   CheckDNSUsage
}

type CheckHTTPUsage struct {
	Total  int64
	PerOrg map[string]int64
}

type CheckHTTPSUsage struct {
	Total  int64
	PerOrg map[string]int64
}
type CheckPINGUsage struct {
	Total  int64
	PerOrg map[string]int64
}
type CheckDNSUsage struct {
	Total  int64
	PerOrg map[string]int64
}

func NewUsage() *Usage {
	return &Usage{
		Endpoints: EndpointUsage{
			PerOrg: make(map[string]int64),
		},
		Probes: ProbeUsage{
			PerOrg: make(map[string]int64),
		},
		Checks: CheckUsage{
			HTTP: CheckHTTPUsage{
				PerOrg: make(map[string]int64),
			},
			HTTPS: CheckHTTPSUsage{
				PerOrg: make(map[string]int64),
			},
			PING: CheckPINGUsage{
				PerOrg: make(map[string]int64),
			},
			DNS: CheckDNSUsage{
				PerOrg: make(map[string]int64),
			},
		},
	}
}

type BillingUsage struct {
	OrgId           int64
	ChecksPerMinute float64
}
