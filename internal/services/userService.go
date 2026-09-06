package services

import (
	"database/sql"
	"errors"

	"github.com/dhruv15803/url-shortener-service/internal/models"
	"github.com/dhruv15803/url-shortener-service/internal/repositories"
)

// ErrUserNotFound is returned when a session references a user that no longer
// exists - e.g. a token issued before the account was removed.
var ErrUserNotFound = errors.New("user not found")

// UserService holds the full repository aggregate (not just repository.Users)
// so its business logic can reach into other domains' repositories when needed.
type UserService struct {
	repository *repositories.Repository
}

func NewUserService(repository *repositories.Repository) *UserService {
	return &UserService{
		repository: repository,
	}
}

func (s *UserService) DeleteUserByID(id int) error {
	return s.repository.Users.DeleteUserByID(id)
}

func (s *UserService) GetUserByID(id int) (*models.User, error) {
	user, err := s.repository.Users.GetUserByID(id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}

	return user, nil
}
