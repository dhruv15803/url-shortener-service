package repositories

import (
	"github.com/dhruv15803/url-shortener-service/internal/models"
	"github.com/jmoiron/sqlx"
)

type Repository struct {
	Users IUserRepository
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{
		Users: NewUserRepository(db),
	}
}

type IUserRepository interface {
	DeleteUserByID(id int) error
	UpsertGoogleUser(user *models.User) (*models.User, error)
}
