package repositories

import (
	"github.com/dhruv15803/url-shortener-service/internal/models"
	"github.com/jmoiron/sqlx"
)

type ClickRepository struct {
	db *sqlx.DB
}

func NewClickRepository(db *sqlx.DB) *ClickRepository {
	return &ClickRepository{
		db: db,
	}
}

func (c *ClickRepository) Create(event models.ClickEvent) error {
	query := `
		INSERT INTO clicks (short_url_id, clicked_at, ip, user_agent, referrer, country, city, region, device, browser, os)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	_, err := c.db.Exec(query,
		event.ShortURLID,
		event.ClickedAt,
		event.IP,
		event.UserAgent,
		event.Referrer,
		event.Country,
		event.City,
		event.Region,
		event.Device,
		event.Browser,
		event.OS,
	)

	return err
}
