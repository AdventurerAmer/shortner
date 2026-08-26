package config

import "time"

type Constants struct {
	AnalyticClicksCacheTTL time.Duration `koanf:"analyticClicksCacheTTL"`
}
