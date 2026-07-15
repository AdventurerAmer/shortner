package domain

import "time"

type Topic string

const (
	ClicksTopic      Topic = "clicks"
	ClicksBatchTopic Topic = "clicksBatch"
)

const ClickEventTimeout = 2 * time.Second

type ClickEvent struct {
	Alias     string    `json:"alias"`
	Timestamp time.Time `json:"timestamp"`
}

const ClicksBatchEventTimeout = 2 * time.Second

type ClicksBatchEvent struct {
	UUId    string   `json:"uuid"`
	Aliases []string `json:"aliases"`
	Clicks  []int    `json:"clicks"`
}
