package models

import "time"

type Click struct {
	ID         int       `db:"id" json:"id"`
	ShortURLID int       `db:"short_url_id" json:"short_url_id"`
	ClickedAt  time.Time `db:"clicked_at" json:"clicked_at"`
	IP         *string   `db:"ip" json:"ip,omitempty"`
	UserAgent  *string   `db:"user_agent" json:"user_agent,omitempty"`
	Referrer   *string   `db:"referrer" json:"referrer,omitempty"`
	Country    *string   `db:"country" json:"country,omitempty"`
	City       *string   `db:"city" json:"city,omitempty"`
	Region     *string   `db:"region" json:"region,omitempty"`
	Device     *string   `db:"device" json:"device,omitempty"`
	Browser    *string   `db:"browser" json:"browser,omitempty"`
	OS         *string   `db:"os" json:"os,omitempty"`
}

// ClickEvent is the enriched, serializable payload put on the queue by the api
// and consumed by the click worker. It is a Click without the db-assigned id.
type ClickEvent struct {
	ShortURLID int       `json:"short_url_id"`
	ClickedAt  time.Time `json:"clicked_at"`
	IP         *string   `json:"ip,omitempty"`
	UserAgent  *string   `json:"user_agent,omitempty"`
	Referrer   *string   `json:"referrer,omitempty"`
	Country    *string   `json:"country,omitempty"`
	City       *string   `json:"city,omitempty"`
	Region     *string   `json:"region,omitempty"`
	Device     *string   `json:"device,omitempty"`
	Browser    *string   `json:"browser,omitempty"`
	OS         *string   `json:"os,omitempty"`
}
