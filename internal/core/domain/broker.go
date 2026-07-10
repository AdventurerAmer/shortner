package domain

import "time"

type Topic string

const (
	ClicksTopic Topic = "clicks"
)

type ClickEvent struct {
	Alias     string    `json:"alias"`
	Timestamp time.Time `json:"timestamp"`
}
