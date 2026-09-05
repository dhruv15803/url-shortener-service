package repositories

import (
	"encoding/base64"
	"strconv"
	"time"

	"github.com/dhruv15803/url-shortener-service/internal/models"
	"github.com/jmoiron/sqlx"
)

// ShortURLWithDestination is a short url joined to the url it points at, so
// listing campaigns is a single query instead of one lookup per row.
type ShortURLWithDestination struct {
	models.ShortURL
	DestinationURL string `db:"destination_url"`
}

type ShortURLRepository struct {
	db *sqlx.DB
}

func NewShortURLRepository(db *sqlx.DB) *ShortURLRepository {
	return &ShortURLRepository{
		db: db,
	}
}

func (s *ShortURLRepository) GetByShortCode(shortCode string) (*models.ShortURL, error) {
	query := `SELECT * FROM short_urls WHERE short_code = $1`

	var shortURL models.ShortURL
	if err := s.db.Get(&shortURL, query, shortCode); err != nil {
		return nil, err
	}

	return &shortURL, nil
}

func (s *ShortURLRepository) ListByUserID(userID int, limit int, offset int) ([]*ShortURLWithDestination, error) {
	query := `
		SELECT s.*, d.destination_url
		FROM short_urls s
		JOIN destination_urls d ON d.id = s.destination_id
		WHERE d.user_id = $1
		ORDER BY s.created_at DESC, s.id DESC
		LIMIT $2 OFFSET $3
	`

	shortURLs := []*ShortURLWithDestination{}
	if err := s.db.Select(&shortURLs, query, userID, limit, offset); err != nil {
		return nil, err
	}

	return shortURLs, nil
}

func (s *ShortURLRepository) CountByUserID(userID int) (int, error) {
	query := `
		SELECT count(*)
		FROM short_urls s
		JOIN destination_urls d ON d.id = s.destination_id
		WHERE d.user_id = $1
	`

	var total int
	if err := s.db.Get(&total, query, userID); err != nil {
		return 0, err
	}

	return total, nil
}

// Update writes all four mutable fields at once. The service merges the
// caller's patch onto the current row first, so the sql here stays static.
func (s *ShortURLRepository) Update(id int, name *string, status models.ShortURLStatus, startsAt *time.Time, expiresAt *time.Time) (*models.ShortURL, error) {
	query := `
		UPDATE short_urls
		SET name = $1, status = $2, starts_at = $3, expires_at = $4
		WHERE id = $5
		RETURNING *
	`

	var shortURL models.ShortURL
	if err := s.db.Get(&shortURL, query, name, status, startsAt, expiresAt, id); err != nil {
		return nil, err
	}

	return &shortURL, nil
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
