package services

import (
	"context"

	"github.com/dhruv15803/url-shortener-service/internal/cache"
	"github.com/dhruv15803/url-shortener-service/internal/config"
	"github.com/dhruv15803/url-shortener-service/internal/geoip"
	"github.com/dhruv15803/url-shortener-service/internal/models"
	"github.com/dhruv15803/url-shortener-service/internal/queue"
	"github.com/dhruv15803/url-shortener-service/internal/repositories"
)

type Service struct {
	Users  IUserService
	Auth   IAuthService
	Urls   IUrlService
	Clicks IClickService
}

func NewService(repository *repositories.Repository, cfg *config.Config, clickQueue *queue.ClickQueue, geoipReader *geoip.Reader, shortURLCache *cache.ShortURLCache) *Service {
	return &Service{
		Users:  NewUserService(repository),
		Auth:   NewAuthService(repository, cfg),
		Urls:   NewUrlService(repository, shortURLCache),
		Clicks: NewClickService(repository, clickQueue, geoipReader),
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
	FindOrCreateDestination(ctx context.Context, userID int, input CreateDestinationInput) (*FindOrCreateDestinationResult, error)
	CreateShortURLForDestination(ctx context.Context, userID int, destinationID int, input CreateShortURLInput) (*models.ShortURL, error)
	ResolveShortCode(ctx context.Context, shortCode string) (string, int, error)
	ListCampaigns(userID int, limit int, offset int) ([]*CampaignView, int, error)
	GetCampaign(userID int, shortCode string) (*CampaignView, error)
	UpdateShortURL(ctx context.Context, userID int, shortCode string, input UpdateShortURLInput) (*CampaignView, error)
}

type IClickService interface {
	RecordClick(ctx context.Context, input RecordClickInput) error
	PersistClick(event models.ClickEvent) error
}
