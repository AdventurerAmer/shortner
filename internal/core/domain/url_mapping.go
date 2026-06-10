package domain

import "time"

type URLMapping struct {
	ShortURL  string    `json:"shortURL"`
	LongURL   string    `json:"longURL"`
	CreatedAt time.Time `json:"createdAt"`
}
