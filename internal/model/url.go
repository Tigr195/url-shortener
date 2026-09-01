package model

import "time"

type URL struct {
	ID          int64      `db:"id"`
	ShortCode   string     `db:"short_code"`
	OriginalURL string     `db:"original_url"`
	CreatedAt   time.Time  `db:"created_at"`
	ExpiresAt   *time.Time `db:"expires_at"`
	Clicks      int64      `db:"clicks"`
}

type ShortenRequest struct {
	URL string `json:"url"`
}

type ShortenResponse struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}
