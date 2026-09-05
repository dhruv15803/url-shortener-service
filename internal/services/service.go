package services

import (
	"context"

	"github.com/dhruv15803/url-shortener-service/internal/config"
	"github.com/dhruv15803/url-shortener-service/internal/models"
	"github.com/dhruv15803/url-shortener-service/internal/repositories"
)

type Service struct {
	Users IUserService
	Auth  IAuthService
}

func NewService(repository *repositories.Repository, cfg *config.Config) *Service {
	return &Service{
		Users: NewUserService(repository),
		Auth:  NewAuthService(repository, cfg),
	}
}

type IUserService interface {
	DeleteUserByID(id int) error
}

type IAuthService interface {
	GenerateState() (string, error)
	GoogleAuthCodeURL(state string) string
	HandleGoogleCallback(ctx context.Context, code string) (*models.User, string, error)
}
