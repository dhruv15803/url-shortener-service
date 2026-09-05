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
	Device     *string   `db:"device" json:"device,omitempty"`
	Browser    *string   `db:"browser" json:"browser,omitempty"`
}
