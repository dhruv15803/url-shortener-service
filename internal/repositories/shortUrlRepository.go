package repositories

import (
	"encoding/base64"
	"strconv"
	"time"

	"github.com/dhruv15803/url-shortener-service/internal/models"
	"github.com/jmoiron/sqlx"
)

type ShortURLRepository struct {
	db *sqlx.DB
}

func NewShortURLRepository(db *sqlx.DB) *ShortURLRepository {
	return &ShortURLRepository{
		db: db,
	}
}

func (s *ShortURLRepository) GetByDestinationID(destinationID int) ([]*models.ShortURL, error) {
	query := `SELECT * FROM short_urls WHERE destination_id = $1 ORDER BY created_at DESC`

	shortURLs := []*models.ShortURL{}
	if err := s.db.Select(&shortURLs, query, destinationID); err != nil {
		return nil, err
	}

	return shortURLs, nil
}

// Create reserves the next id from the short_urls sequence, derives the short
// code from that id, and inserts the row with both in a single statement.
// Reserving the id up front is what lets every short url for the same
// destination get its own distinct code.
func (s *ShortURLRepository) Create(destinationID int, name *string, status models.ShortURLStatus, startsAt *time.Time, expiresAt *time.Time) (*models.ShortURL, error) {
	var id int
	if err := s.db.Get(&id, `SELECT nextval('short_urls_id_seq')`); err != nil {
		return nil, err
	}

	shortCode := base64.RawURLEncoding.EncodeToString([]byte(strconv.Itoa(id)))

	query := `
		INSERT INTO short_urls (id, short_code, destination_id, name, status, starts_at, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING *
	`

	var shortURL models.ShortURL
	if err := s.db.Get(&shortURL, query, id, shortCode, destinationID, name, status, startsAt, expiresAt); err != nil {
		return nil, err
	}

	return &shortURL, nil
}
