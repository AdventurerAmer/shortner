package health

import "context"

type LivenessStatus string

const (
	LivenessStatusUp = "up"
)

type LivenessCheck struct {
	Status LivenessStatus `json:"status"`
}

type ReadinessStatus string

const (
	ReadinessStatusReady    ReadinessStatus = "ready"
	ReadinessStatusNotReady ReadinessStatus = "notReady"
)

type Checks map[string]string

type ReadinessCheck struct {
	Status ReadinessStatus `json:"status"`
	Checks Checks          `json:"checks"`
}

type ReadinessHandler func(ctx context.Context, checks Checks) error
