package repositories

import (
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
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

// CampaignFilter narrows a campaign list to what the user asked for. The zero
// value filters nothing.
type CampaignFilter struct {
	// Search is a case-insensitive substring matched against the campaign
	// name, the short code and the destination url. "" means no filter.
	Search string

	// Status is an *effective* status, so it may be scheduled or expired even
	// though neither is ever stored. "" means no filter.
	Status models.ShortURLStatus
}

// effectiveStatusExpr is the sql twin of models.ShortURL.EffectiveStatus.
// Scheduled and expired are not stored - only the columns they are derived
// from are - so filtering on them means recomputing the label here. The two
// implementations must be changed together or the list will disagree with the
// status badge each row renders.
//
// now is passed in rather than using sql now() so that both sides of that pair
// judge a campaign against the same instant.
const effectiveStatusExpr = `
		CASE
			WHEN s.status = 'disabled' THEN 'disabled'
			WHEN s.starts_at IS NOT NULL AND s.starts_at > %[1]s THEN 'scheduled'
			WHEN s.expires_at IS NOT NULL AND s.expires_at < %[1]s THEN 'expired'
			ELSE 'active'
		END`

// likeEscape neutralises the wildcards in a user's search term so that
// searching for "50%" means the literal characters and not "match everything".
var likeEscape = strings.NewReplacer(`\`, `\`, `%`, `\%`, `_`, `\_`)

// campaignConditions builds the WHERE body shared by ListByUserID and
// CountByUserID. Both must apply identical filters or the reported total would
// not match the rows on the page.
func campaignConditions(userID int, filter CampaignFilter, now time.Time) (string, []any) {
	conditions := []string{"d.user_id = $1"}
	args := []any{userID}

	if filter.Search != "" {
		args = append(args, "%"+likeEscape.Replace(filter.Search)+"%")
		// name is nullable, and NULL ILIKE is NULL, so an untitled campaign
		// simply fails that arm instead of matching everything.
		conditions = append(conditions, fmt.Sprintf(
			`(s.name ILIKE $%[1]d ESCAPE '\' OR s.short_code ILIKE $%[1]d ESCAPE '\' OR d.destination_url ILIKE $%[1]d ESCAPE '\')`,
			len(args),
		))
	}

	if filter.Status != "" {
		args = append(args, now)
		nowArg := fmt.Sprintf("$%d", len(args))

		args = append(args, string(filter.Status))
		conditions = append(conditions, fmt.Sprintf(effectiveStatusExpr, nowArg)+fmt.Sprintf(" = $%d", len(args)))
	}

	return strings.Join(conditions, " AND "), args
}

func (s *ShortURLRepository) ListByUserID(userID int, filter CampaignFilter, now time.Time, limit int, offset int) ([]*ShortURLWithDestination, error) {
	where, args := campaignConditions(userID, filter, now)
	args = append(args, limit, offset)

	query := fmt.Sprintf(`
		SELECT s.*, d.destination_url
		FROM short_urls s
		JOIN destination_urls d ON d.id = s.destination_id
		WHERE %s
		ORDER BY s.created_at DESC, s.id DESC
		LIMIT $%d OFFSET $%d
	`, where, len(args)-1, len(args))

	shortURLs := []*ShortURLWithDestination{}
	if err := s.db.Select(&shortURLs, query, args...); err != nil {
		return nil, err
	}

	return shortURLs, nil
}

func (s *ShortURLRepository) CountByUserID(userID int, filter CampaignFilter, now time.Time) (int, error) {
	where, args := campaignConditions(userID, filter, now)

	query := fmt.Sprintf(`
		SELECT count(*)
		FROM short_urls s
		JOIN destination_urls d ON d.id = s.destination_id
		WHERE %s
	`, where)

	var total int
	if err := s.db.Get(&total, query, args...); err != nil {
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
