package models

import (
	"strings"
	"time"
)

// ShortURLStatus has four values, but only two of them are ever stored.
//
// The status column holds the user's intent - active (enabled) or disabled
// (paused). Scheduled and expired are derived from starts_at/expires_at at
// read time by EffectiveStatus, because a stored value would silently go
// stale the moment a campaign's start or end time passed.
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

// EffectiveStatus is the status callers should see. Disabled is a deliberate
// user action and always wins; everything else follows from the schedule.
func (s *ShortURL) EffectiveStatus(now time.Time) ShortURLStatus {
	if s.Status == ShortURLStatusDisabled {
		return ShortURLStatusDisabled
	}

	if s.StartsAt != nil && s.StartsAt.After(now) {
		return ShortURLStatusScheduled
	}

	if s.ExpiresAt != nil && s.ExpiresAt.Before(now) {
		return ShortURLStatusExpired
	}

	return ShortURLStatusActive
}

// IsResolvable reports whether the short url should redirect right now.
func (s *ShortURL) IsResolvable(now time.Time) bool {
	return s.EffectiveStatus(now) == ShortURLStatusActive
}

// ParseShortURLStatus accepts any of the four statuses a campaign can present,
// including scheduled and expired, which are derived rather than stored. It is
// for reading filters off a request - never for deciding what to write.
func ParseShortURLStatus(raw string) (ShortURLStatus, bool) {
	switch ShortURLStatus(strings.ToLower(strings.TrimSpace(raw))) {
	case ShortURLStatusScheduled:
		return ShortURLStatusScheduled, true
	case ShortURLStatusActive:
		return ShortURLStatusActive, true
	case ShortURLStatusDisabled:
		return ShortURLStatusDisabled, true
	case ShortURLStatusExpired:
		return ShortURLStatusExpired, true
	default:
		return "", false
	}
}
