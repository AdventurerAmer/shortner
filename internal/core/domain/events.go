package domain

import (
	"time"
)

type Topic string

const (
	ClicksTopic      Topic = "clicks"
	ClicksBatchTopic Topic = "clicksBatch"
)

type ClickEvent struct {
	Alias     string    `json:"alias"`
	Timestamp time.Time `json:"timestamp"`
}

type ClicksBatchEvent struct {
	UUIds   []string `json:"uuid"`
	Aliases []string `json:"aliases"`
	Clicks  []int    `json:"clicks"`
}
