package models

import "time"

type DestinationURL struct {
	ID             int       `db:"id" json:"id"`
	DestinationURL string    `db:"destination_url" json:"destination_url"`
	UserID         int       `db:"user_id" json:"user_id"`
	CreatedAt      time.Time `db:"created_at" json:"created_at"`
	UpdatedAt      time.Time `db:"updated_at" json:"updated_at"`
}
