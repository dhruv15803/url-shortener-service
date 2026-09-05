package models

import "time"

type User struct {
	ID         int        `db:"id" json:"id"`
	Email      string     `db:"email" json:"email"`
	Password   *string    `db:"password" json:"-"`
	GoogleID   *string    `db:"google_id" json:"-"`
	ImageURL   *string    `db:"image_url" json:"image_url,omitempty"`
	Username   string     `db:"username" json:"username"`
	IsVerified bool       `db:"is_verified" json:"is_verified"`
	VerifiedAt *time.Time `db:"verified_at" json:"verified_at,omitempty"`
	CreatedAt  time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt  time.Time  `db:"updated_at" json:"updated_at"`
}
