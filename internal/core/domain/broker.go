package domain

import (
	"time"

	"github.com/sony/gobreaker/v2"
)

const SendEventDefaultTimeout = 2 * time.Second
const SendEventDefaultRetryAttempts = 10

var SendEventCircuitBreaker = gobreaker.NewCircuitBreaker[[]byte](gobreaker.Settings{
	Name:        "sendEvent",
	Timeout:     30 * time.Second, // Time in Open state before Half-Open
	MaxRequests: 5,                // Requests allowed in Half-Open
	Interval:    60 * time.Second, // Clear counts periodically in Closed
	ReadyToTrip: func(counts gobreaker.Counts) bool {
		return counts.ConsecutiveFailures > 5
	},
	IsSuccessful: func(err error) bool {
		return err == nil
	},
})
