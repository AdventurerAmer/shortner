package domain

import "time"

type Analytic struct {
	Alias     string    `json:"alias"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	Clicks    int       `json:"clicks"`
	Version   int       `json:"version"`
}
