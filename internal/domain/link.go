package domain

import "time"

type Link struct {
	ID        int64
	Code      string
	URL       string
	CreatedAt time.Time
}
