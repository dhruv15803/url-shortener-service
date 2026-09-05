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
	Urls  IUrlService
}

func NewService(repository *repositories.Repository, cfg *config.Config) *Service {
	return &Service{
		Users: NewUserService(repository),
		Auth:  NewAuthService(repository, cfg),
		Urls:  NewUrlService(repository),
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

type IUrlService interface {
	FindOrCreateDestination(userID int, input CreateDestinationInput) (*FindOrCreateDestinationResult, error)
	CreateShortURLForDestination(userID int, destinationID int, input CreateShortURLInput) (*models.ShortURL, error)
}
