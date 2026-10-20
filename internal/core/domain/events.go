package domain

import (
	"time"
)

type Topic string

const (
	ClicksTopic Topic = "clicks"
)

type ClickEvent struct {
	Id        string    `json:"id"`
	Alias     string    `json:"alias"`
	Timestamp time.Time `json:"timestamp"`
}
