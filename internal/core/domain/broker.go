package domain

import "time"

type Topic string

const (
	ClicksTopic          Topic = "clicks"
	CollectedClicksTopic Topic = "collectedClicks"
)

const ClickEventTimeout = 2 * time.Second

type ClickEvent struct {
	Alias     string    `json:"alias"`
	Timestamp time.Time `json:"timestamp"`
}

const CollectedClicksEventTimeout = 2 * time.Second

type CollectedClicksEvent struct {
	Alias  string `json:"alias"`
	Clicks int    `json:"clicks"`
}
