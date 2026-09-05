package models

import "time"

type ShortURLStatus string

const (
	ShortURLStatusScheduled ShortURLStatus = "scheduled"
	ShortURLStatusActive    ShortURLStatus = "active"
	ShortURLStatusDisabled  ShortURLStatus = "disabled"
	ShortURLStatusExpired   ShortURLStatus = "expired"
)

type ShortURL struct {
	ID            int            `db:"id" json:"id"`
	ShortCode     string         `db:"short_code" json:"short_code"`
	DestinationID int            `db:"destination_id" json:"destination_id"`
	Name          *string        `db:"name" json:"name,omitempty"`
	Status        ShortURLStatus `db:"status" json:"status"`
	StartsAt      *time.Time     `db:"starts_at" json:"starts_at,omitempty"`
	ExpiresAt     *time.Time     `db:"expires_at" json:"expires_at,omitempty"`
	CreatedAt     time.Time      `db:"created_at" json:"created_at"`
}
