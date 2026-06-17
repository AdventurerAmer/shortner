package domain

import "time"

type URLMapping struct {
	Alias     string    `json:"alias"`
	LongURL   string    `json:"longURL"`
	CreatedAt time.Time `json:"createdAt"`
	UserId    string    `json:"userId"`
}
