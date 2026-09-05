package services

import (
	"database/sql"
	"errors"
	"time"

	"github.com/dhruv15803/url-shortener-service/internal/models"
	"github.com/dhruv15803/url-shortener-service/internal/repositories"
)

// ErrDestinationNotFound is returned when a destination does not exist, or
// exists but belongs to another user - callers map both to a 404 so the
// existence of other users' destinations isn't leaked.
var ErrDestinationNotFound = errors.New("destination not found")

type CreateDestinationInput struct {
	DestinationURL string
	CampaignName   *string
	StartsAt       *time.Time
	ExpiresAt      *time.Time
}

type CreateShortURLInput struct {
	CampaignName *string
	StartsAt     *time.Time
	ExpiresAt    *time.Time
}

type FindOrCreateDestinationResult struct {
	Destination       *models.DestinationURL
	Created           bool
	ExistingShortURLs []*models.ShortURL
	NewShortURL       *models.ShortURL
}

type UrlService struct {
	repository *repositories.Repository
}

func NewUrlService(repository *repositories.Repository) *UrlService {
	return &UrlService{
		repository: repository,
	}
}

func (s *UrlService) FindOrCreateDestination(userID int, input CreateDestinationInput) (*FindOrCreateDestinationResult, error) {
	destination, err := s.repository.DestinationURLs.GetByUserIDAndURL(userID, input.DestinationURL)
	if err == nil {
		return s.existingDestinationResult(destination)
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	destination, err = s.repository.DestinationURLs.Create(userID, input.DestinationURL)
	if errors.Is(err, sql.ErrNoRows) {
		// A concurrent request created this destination first - return what it created.
		destination, err = s.repository.DestinationURLs.GetByUserIDAndURL(userID, input.DestinationURL)
		if err != nil {
			return nil, err
		}
		return s.existingDestinationResult(destination)
	}
	if err != nil {
		return nil, err
	}

	shortURL, err := s.repository.ShortURLs.Create(destination.ID, input.CampaignName, resolveInitialStatus(input.StartsAt), input.StartsAt, input.ExpiresAt)
	if err != nil {
		return nil, err
	}

	return &FindOrCreateDestinationResult{
		Destination: destination,
		Created:     true,
		NewShortURL: shortURL,
	}, nil
}

func (s *UrlService) CreateShortURLForDestination(userID int, destinationID int, input CreateShortURLInput) (*models.ShortURL, error) {
	destination, err := s.repository.DestinationURLs.GetByID(destinationID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrDestinationNotFound
	}
	if err != nil {
		return nil, err
	}

	if destination.UserID != userID {
		return nil, ErrDestinationNotFound
	}

	return s.repository.ShortURLs.Create(destination.ID, input.CampaignName, resolveInitialStatus(input.StartsAt), input.StartsAt, input.ExpiresAt)
}

func (s *UrlService) existingDestinationResult(destination *models.DestinationURL) (*FindOrCreateDestinationResult, error) {
	shortURLs, err := s.repository.ShortURLs.GetByDestinationID(destination.ID)
	if err != nil {
		return nil, err
	}

	return &FindOrCreateDestinationResult{
		Destination:       destination,
		Created:           false,
		ExistingShortURLs: shortURLs,
	}, nil
}

func resolveInitialStatus(startsAt *time.Time) models.ShortURLStatus {
	if startsAt != nil && startsAt.After(time.Now()) {
		return models.ShortURLStatusScheduled
	}
	return models.ShortURLStatusActive
}
