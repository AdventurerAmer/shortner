package domain

import "time"

type Topic string

const (
	ClicksTopic Topic = "clicks"
)

const ClickEventTimeout = 2 * time.Second

type ClickEvent struct {
	Alias     string    `json:"alias"`
	Timestamp time.Time `json:"timestamp"`
}
