package repositories

import (
	"fmt"
	"time"

	"github.com/dhruv15803/url-shortener-service/internal/models"
	"github.com/jmoiron/sqlx"
)

// ClickDimension is a column clicks can be grouped by. It is a closed set of
// typed constants rather than a string, so a user-supplied value can never
// reach the sql - the only way in is through ParseClickDimension.
type ClickDimension string

const (
	DimensionCountry ClickDimension = "country"
	DimensionCity    ClickDimension = "city"
	DimensionRegion  ClickDimension = "region"
	DimensionDevice  ClickDimension = "device"
	DimensionBrowser ClickDimension = "browser"
	DimensionOS      ClickDimension = "os"
)

// UnknownDimensionValue labels rows whose dimension is null. Most clicks have
// no geo (private ips) or no parsed os, so these need to be visible and
// counted rather than silently dropped.
const UnknownDimensionValue = "unknown"

func ParseClickDimension(raw string) (ClickDimension, bool) {
	switch ClickDimension(raw) {
	case DimensionCountry, DimensionCity, DimensionRegion,
		DimensionDevice, DimensionBrowser, DimensionOS:
		return ClickDimension(raw), true
	default:
		return "", false
	}
}

func AllClickDimensions() []string {
	return []string{
		string(DimensionCountry), string(DimensionCity), string(DimensionRegion),
		string(DimensionDevice), string(DimensionBrowser), string(DimensionOS),
	}
}

// column maps a dimension to a literal column name. Unreachable with an
// invalid value, but it returns ok rather than panicking so a future caller
// that skips ParseClickDimension fails safely.
func (d ClickDimension) column() (string, bool) {
	switch d {
	case DimensionCountry:
		return "country", true
	case DimensionCity:
		return "city", true
	case DimensionRegion:
		return "region", true
	case DimensionDevice:
		return "device", true
	case DimensionBrowser:
		return "browser", true
	case DimensionOS:
		return "os", true
	default:
		return "", false
	}
}

// BucketSize is the granularity of the click time series, again a closed set
// so the date_trunc unit is never caller-supplied.
type BucketSize string

const (
	BucketHour BucketSize = "hour"
	BucketDay  BucketSize = "day"
)

func (b BucketSize) unit() (string, bool) {
	switch b {
	case BucketHour:
		return "hour", true
	case BucketDay:
		return "day", true
	default:
		return "", false
	}
}

type DimensionCount struct {
	Value  string `db:"value"`
	Clicks int    `db:"clicks"`
}

type BucketCount struct {
	Bucket time.Time `db:"bucket"`
	Clicks int       `db:"clicks"`
}

type ClickRepository struct {
	db *sqlx.DB
}

func NewClickRepository(db *sqlx.DB) *ClickRepository {
	return &ClickRepository{
		db: db,
	}
}

func (c *ClickRepository) CountByShortURL(shortURLID int, start time.Time, end time.Time) (int, error) {
	query := `
		SELECT count(*) FROM clicks
		WHERE short_url_id = $1 AND clicked_at >= $2 AND clicked_at <= $3
	`

	var total int
	if err := c.db.Get(&total, query, shortURLID, start, end); err != nil {
		return 0, err
	}

	return total, nil
}

// GroupByDimension returns one page of a dimension's values, busiest first.
// The column name comes from the typed dimension, never from input.
func (c *ClickRepository) GroupByDimension(shortURLID int, dimension ClickDimension, start time.Time, end time.Time, limit int, offset int) ([]DimensionCount, error) {
	column, ok := dimension.column()
	if !ok {
		return nil, fmt.Errorf("unsupported click dimension: %q", dimension)
	}

	// The value tiebreak is what makes paging safe: with equal counts and no
	// secondary sort, rows could reorder between pages and be duplicated or
	// skipped entirely.
	query := fmt.Sprintf(`
		SELECT COALESCE(%s, '%s') AS value, count(*) AS clicks
		FROM clicks
		WHERE short_url_id = $1 AND clicked_at >= $2 AND clicked_at <= $3
		GROUP BY 1
		ORDER BY clicks DESC, value ASC
		LIMIT $4 OFFSET $5
	`, column, UnknownDimensionValue)

	counts := []DimensionCount{}
	if err := c.db.Select(&counts, query, shortURLID, start, end, limit, offset); err != nil {
		return nil, err
	}

	return counts, nil
}

// CountDimensionGroups is how many distinct values a dimension has in the
// range - the row count, not the click count, so the client can work out how
// many pages there are.
func (c *ClickRepository) CountDimensionGroups(shortURLID int, dimension ClickDimension, start time.Time, end time.Time) (int, error) {
	column, ok := dimension.column()
	if !ok {
		return 0, fmt.Errorf("unsupported click dimension: %q", dimension)
	}

	query := fmt.Sprintf(`
		SELECT count(DISTINCT COALESCE(%s, '%s'))
		FROM clicks
		WHERE short_url_id = $1 AND clicked_at >= $2 AND clicked_at <= $3
	`, column, UnknownDimensionValue)

	var total int
	if err := c.db.Get(&total, query, shortURLID, start, end); err != nil {
		return 0, err
	}

	return total, nil
}

func (c *ClickRepository) SeriesByBucket(shortURLID int, bucket BucketSize, start time.Time, end time.Time) ([]BucketCount, error) {
	unit, ok := bucket.unit()
	if !ok {
		return nil, fmt.Errorf("unsupported bucket size: %q", bucket)
	}

	query := fmt.Sprintf(`
		SELECT date_trunc('%s', clicked_at) AS bucket, count(*) AS clicks
		FROM clicks
		WHERE short_url_id = $1 AND clicked_at >= $2 AND clicked_at <= $3
		GROUP BY 1
		ORDER BY bucket ASC
	`, unit)

	buckets := []BucketCount{}
	if err := c.db.Select(&buckets, query, shortURLID, start, end); err != nil {
		return nil, err
	}

	return buckets, nil
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
