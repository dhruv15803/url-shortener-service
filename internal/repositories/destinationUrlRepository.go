package repositories

import (
	"github.com/dhruv15803/url-shortener-service/internal/models"
	"github.com/jmoiron/sqlx"
)

type DestinationURLRepository struct {
	db *sqlx.DB
}

func NewDestinationURLRepository(db *sqlx.DB) *DestinationURLRepository {
	return &DestinationURLRepository{
		db: db,
	}
}

func (d *DestinationURLRepository) GetByUserIDAndURL(userID int, destinationURL string) (*models.DestinationURL, error) {
	query := `SELECT * FROM destination_urls WHERE user_id = $1 AND destination_url = $2`

	var destination models.DestinationURL
	if err := d.db.Get(&destination, query, userID, destinationURL); err != nil {
		return nil, err
	}

	return &destination, nil
}

func (d *DestinationURLRepository) GetByID(id int) (*models.DestinationURL, error) {
	query := `SELECT * FROM destination_urls WHERE id = $1`

	var destination models.DestinationURL
	if err := d.db.Get(&destination, query, id); err != nil {
		return nil, err
	}

	return &destination, nil
}

// Create inserts a new destination for the user. When a concurrent request
// has already inserted the same destination, no row is returned and the
// error is sql.ErrNoRows - the caller should then re-fetch the existing row.
func (d *DestinationURLRepository) Create(userID int, destinationURL string) (*models.DestinationURL, error) {
	query := `
		INSERT INTO destination_urls (user_id, destination_url)
		VALUES ($1, $2)
		ON CONFLICT (user_id, destination_url) DO NOTHING
		RETURNING *
	`

	var destination models.DestinationURL
	if err := d.db.Get(&destination, query, userID, destinationURL); err != nil {
		return nil, err
	}

	return &destination, nil
}
