package analytics

import "time"

// URLCreatedEvent represents an event emitted when a URL is shortened.
type URLCreatedEvent struct {
	Code        string    `json:"code"`
	OriginalURL string    `json:"originalUrl"`
	URLHash     string    `json:"urlHash,omitempty"`
	Strategy    string    `json:"strategy"`
	CreatedAt   time.Time `json:"createdAt"`
	ClientIP    string    `json:"clientIp"`
	UserAgent   string    `json:"userAgent"`
}

// URLAccessedEvent represents an event emitted when a short URL is accessed.
type URLAccessedEvent struct {
	Code       string    `json:"code"`
	AccessedAt time.Time `json:"accessedAt"`
	ClientIP   string    `json:"clientIp"`
	UserAgent  string    `json:"userAgent"`
	Referrer   string    `json:"referrer,omitempty"`
}
