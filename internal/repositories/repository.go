package repositories

import (
	"time"

	"github.com/dhruv15803/url-shortener-service/internal/models"
	"github.com/jmoiron/sqlx"
)

type Repository struct {
	Users           IUserRepository
	DestinationURLs IDestinationURLRepository
	ShortURLs       IShortURLRepository
	Clicks          IClickRepository
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{
		Users:           NewUserRepository(db),
		DestinationURLs: NewDestinationURLRepository(db),
		ShortURLs:       NewShortURLRepository(db),
		Clicks:          NewClickRepository(db),
	}
}

type IUserRepository interface {
	DeleteUserByID(id int) error
	UpsertGoogleUser(user *models.User) (*models.User, error)
}

type IDestinationURLRepository interface {
	GetByUserIDAndURL(userID int, destinationURL string) (*models.DestinationURL, error)
	GetByID(id int) (*models.DestinationURL, error)
	Create(userID int, destinationURL string) (*models.DestinationURL, error)
}

type IShortURLRepository interface {
	GetByShortCode(shortCode string) (*models.ShortURL, error)
	GetByDestinationID(destinationID int) ([]*models.ShortURL, error)
	ListByUserID(userID int, limit int, offset int) ([]*ShortURLWithDestination, error)
	CountByUserID(userID int) (int, error)
	Create(destinationID int, name *string, status models.ShortURLStatus, startsAt *time.Time, expiresAt *time.Time) (*models.ShortURL, error)
	Update(id int, name *string, status models.ShortURLStatus, startsAt *time.Time, expiresAt *time.Time) (*models.ShortURL, error)
}

type IClickRepository interface {
	Create(event models.ClickEvent) error
}
