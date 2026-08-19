package billing

import "sort"

type Plan struct {
	Key             string
	Name            string
	PriceUSDMinor   int64
	PriceNGNMinor   int64
	MaxTunnels      int
	MaxDomains      int
	MaxMembers      int
	MaxConnections  int
	BandwidthBytes  int64
	RetentionDays   int
	CustomDomains   bool
	PrioritySupport bool
}

var plans = map[string]Plan{
	"free": {
		Key: "free", Name: "Free", MaxTunnels: 2, MaxDomains: 0, MaxMembers: 1, MaxConnections: 10,
		BandwidthBytes: 2 * 1024 * 1024 * 1024, RetentionDays: 3,
	},
	"link": {
		Key: "link", Name: "Link", PriceUSDMinor: 700, PriceNGNMinor: 1_000_000,
		MaxTunnels: 3, MaxDomains: 1, MaxMembers: 3, MaxConnections: 50, BandwidthBytes: 25 * 1024 * 1024 * 1024,
		RetentionDays: 14, CustomDomains: true,
	},
	"route": {
		Key: "route", Name: "Route", PriceUSDMinor: 1500, PriceNGNMinor: 2_100_000,
		MaxTunnels: 5, MaxDomains: 5, MaxMembers: 5, MaxConnections: 100, BandwidthBytes: 100 * 1024 * 1024 * 1024,
		RetentionDays: 30, CustomDomains: true, PrioritySupport: true,
	},
	"edge": {
		Key: "edge", Name: "Edge", PriceUSDMinor: 12_000, PriceNGNMinor: 17_000_000,
		MaxTunnels: 20, MaxDomains: 25, MaxMembers: -1, MaxConnections: 500, BandwidthBytes: 1024 * 1024 * 1024 * 1024,
		RetentionDays: 90, CustomDomains: true, PrioritySupport: true,
	},
}

const annualPeriods = 10

func annualPrice(monthlyMinor int64) int64 { return monthlyMinor * annualPeriods }

func PlanByKey(key string) (Plan, bool) {
	plan, ok := plans[key]
	return plan, ok
}

func AllPlans() []Plan {
	result := make([]Plan, 0, len(plans))
	keys := make([]string, 0, len(plans))

	for key := range plans {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	for _, key := range keys {
		result = append(result, plans[key])
	}

	return result
}
