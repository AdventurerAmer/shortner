package domain

import (
	"time"

	"github.com/sony/gobreaker/v2"
)

const KafkaTimeout = 2 * time.Second
const KafkaRetryAttempts = 10

var KafkaCB = gobreaker.NewCircuitBreaker[[]byte](gobreaker.Settings{
	Name:        "kafka",
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
