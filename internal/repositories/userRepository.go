package repositories

import (
	"github.com/dhruv15803/url-shortener-service/internal/models"
	"github.com/jmoiron/sqlx"
)

type UserRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

func (u *UserRepository) DeleteUserByID(id int) error {
	return nil
}

// UpsertGoogleUser inserts a new Google-authenticated user, or refreshes the
// email/image_url of an existing one matched by google_id, in a single
// atomic statement.
func (u *UserRepository) UpsertGoogleUser(user *models.User) (*models.User, error) {
	query := `
		INSERT INTO users (email, google_id, username, image_url, is_verified, verified_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (google_id) DO UPDATE
		SET email = EXCLUDED.email, image_url = EXCLUDED.image_url, updated_at = NOW()
		RETURNING *
	`

	var result models.User
	if err := u.db.Get(&result, query, user.Email, user.GoogleID, user.Username, user.ImageURL, user.IsVerified, user.VerifiedAt); err != nil {
		return nil, err
	}

	return &result, nil
}
