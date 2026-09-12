package services

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"strings"
	"time"

	"github.com/dhruv15803/url-shortener-service/internal/cache"
	"github.com/dhruv15803/url-shortener-service/internal/models"
	"github.com/dhruv15803/url-shortener-service/internal/optional"
	"github.com/dhruv15803/url-shortener-service/internal/repositories"
)

const maxCampaignNameLength = 255

// ErrDestinationNotFound is returned when a destination does not exist, or
// exists but belongs to another user - callers map both to a 404 so the
// existence of other users' destinations isn't leaked.
var ErrDestinationNotFound = errors.New("destination not found")

// ErrShortURLNotResolvable covers every reason a code can't redirect: unknown
// code, disabled, expired, or not started yet. They're deliberately
// indistinguishable to the caller so a 404 doesn't leak which one it was.
var ErrShortURLNotResolvable = errors.New("short url not resolvable")

// ErrShortURLNotFound is returned when a short code doesn't exist or belongs
// to another user - both map to 404 so ownership isn't leaked.
var ErrShortURLNotFound = errors.New("short url not found")

// ErrInvalidUpdate is returned when a patch would leave the campaign in an
// invalid state. Handlers map it to 400 and surface the message.
type ErrInvalidUpdate struct {
	Message string
}

func (e *ErrInvalidUpdate) Error() string {
	return e.Message
}

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

// UpdateShortURLInput is a patch: an absent field is left alone, an explicit
// null clears it. Status carries the user's intent only ("ACTIVE"/"DISABLED").
type UpdateShortURLInput struct {
	Name      optional.Optional[string]
	Status    optional.Optional[string]
	StartsAt  optional.Optional[time.Time]
	ExpiresAt optional.Optional[time.Time]
}

// CampaignView is a short url plus the destination it points at, with the
// status already derived. It is what every campaign-facing endpoint returns.
type CampaignView struct {
	ShortURL       *models.ShortURL
	DestinationURL string
	Status         models.ShortURLStatus
}

func newCampaignView(shortURL *models.ShortURL, destinationURL string) *CampaignView {
	return newCampaignViewAt(shortURL, destinationURL, time.Now())
}

// newCampaignViewAt derives the status against a caller-supplied instant, so a
// batch of views built together all agree on what "now" was.
func newCampaignViewAt(shortURL *models.ShortURL, destinationURL string, now time.Time) *CampaignView {
	return &CampaignView{
		ShortURL:       shortURL,
		DestinationURL: destinationURL,
		Status:         shortURL.EffectiveStatus(now),
	}
}

type UrlService struct {
	repository *repositories.Repository
	cache      *cache.ShortURLCache
}

func NewUrlService(repository *repositories.Repository, shortURLCache *cache.ShortURLCache) *UrlService {
	return &UrlService{
		repository: repository,
		cache:      shortURLCache,
	}
}

func (s *UrlService) FindOrCreateDestination(ctx context.Context, userID int, input CreateDestinationInput) (*FindOrCreateDestinationResult, error) {
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

	shortURL, err := s.repository.ShortURLs.Create(destination.ID, input.CampaignName, models.ShortURLStatusActive, input.StartsAt, input.ExpiresAt)
	if err != nil {
		return nil, err
	}

	s.warmCache(ctx, shortURL, destination.DestinationURL)

	return &FindOrCreateDestinationResult{
		Destination: destination,
		Created:     true,
		NewShortURL: shortURL,
	}, nil
}

func (s *UrlService) CreateShortURLForDestination(ctx context.Context, userID int, destinationID int, input CreateShortURLInput) (*models.ShortURL, error) {
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

	shortURL, err := s.repository.ShortURLs.Create(destination.ID, input.CampaignName, models.ShortURLStatusActive, input.StartsAt, input.ExpiresAt)
	if err != nil {
		return nil, err
	}

	s.warmCache(ctx, shortURL, destination.DestinationURL)

	return shortURL, nil
}

// ListCampaigns returns one filtered page plus the total matching that same
// filter, so the caller can page through it.
//
// now is read once and used for the status filter, the count and every view,
// so a campaign can't be selected as active by the query and then rendered as
// expired a moment later.
func (s *UrlService) ListCampaigns(userID int, filter repositories.CampaignFilter, limit int, offset int) ([]*CampaignView, int, error) {
	now := time.Now()

	rows, err := s.repository.ShortURLs.ListByUserID(userID, filter, now, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	total, err := s.repository.ShortURLs.CountByUserID(userID, filter, now)
	if err != nil {
		return nil, 0, err
	}

	campaigns := make([]*CampaignView, 0, len(rows))
	for _, row := range rows {
		shortURL := row.ShortURL
		campaigns = append(campaigns, newCampaignViewAt(&shortURL, row.DestinationURL, now))
	}

	return campaigns, total, nil
}

func (s *UrlService) GetCampaign(userID int, shortCode string) (*CampaignView, error) {
	shortURL, destination, err := s.ownedShortURL(userID, shortCode)
	if err != nil {
		return nil, err
	}

	return newCampaignView(shortURL, destination.DestinationURL), nil
}

// UpdateShortURL applies a patch to a campaign and refreshes the redirect
// cache so a disable or reschedule takes effect immediately rather than
// lingering until the cached entry expires.
func (s *UrlService) UpdateShortURL(ctx context.Context, userID int, shortCode string, input UpdateShortURLInput) (*CampaignView, error) {
	shortURL, destination, err := s.ownedShortURL(userID, shortCode)
	if err != nil {
		return nil, err
	}

	name := shortURL.Name
	if input.Name.Set() {
		name = input.Name.Value
	} else if input.Name.Cleared() {
		name = nil
	}

	status := shortURL.Status
	if input.Status.Set() {
		parsed, err := parseStatusIntent(*input.Status.Value)
		if err != nil {
			return nil, err
		}
		status = parsed
	} else if input.Status.Cleared() {
		return nil, &ErrInvalidUpdate{Message: "status cannot be null"}
	}

	startsAt := shortURL.StartsAt
	if input.StartsAt.Set() {
		startsAt = input.StartsAt.Value
	} else if input.StartsAt.Cleared() {
		startsAt = nil
	}

	expiresAt := shortURL.ExpiresAt
	if input.ExpiresAt.Set() {
		expiresAt = input.ExpiresAt.Value
	} else if input.ExpiresAt.Cleared() {
		expiresAt = nil
	}

	// Validate the merged result, not the current row, so that one request can
	// legitimately re-enable a campaign and extend its expiry together.
	if name != nil && len(*name) > maxCampaignNameLength {
		return nil, &ErrInvalidUpdate{Message: "name must be 255 characters or fewer"}
	}

	if startsAt != nil && expiresAt != nil && !expiresAt.After(*startsAt) {
		return nil, &ErrInvalidUpdate{Message: "expires_at must be after starts_at"}
	}

	updated, err := s.repository.ShortURLs.Update(shortURL.ID, name, status, startsAt, expiresAt)
	if err != nil {
		return nil, err
	}

	s.warmCache(ctx, updated, destination.DestinationURL)

	return newCampaignView(updated, destination.DestinationURL), nil
}

// ownedShortURL loads a short url and its destination, treating "doesn't
// exist" and "belongs to someone else" identically.
func (s *UrlService) ownedShortURL(userID int, shortCode string) (*models.ShortURL, *models.DestinationURL, error) {
	shortURL, err := s.repository.ShortURLs.GetByShortCode(shortCode)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil, ErrShortURLNotFound
	}
	if err != nil {
		return nil, nil, err
	}

	destination, err := s.repository.DestinationURLs.GetByID(shortURL.DestinationID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil, ErrShortURLNotFound
	}
	if err != nil {
		return nil, nil, err
	}

	if destination.UserID != userID {
		return nil, nil, ErrShortURLNotFound
	}

	return shortURL, destination, nil
}

// parseStatusIntent accepts only the two statuses a user can actually set.
// Scheduled and expired are computed from the campaign's dates.
func parseStatusIntent(raw string) (models.ShortURLStatus, error) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case string(models.ShortURLStatusActive):
		return models.ShortURLStatusActive, nil
	case string(models.ShortURLStatusDisabled):
		return models.ShortURLStatusDisabled, nil
	case string(models.ShortURLStatusScheduled), string(models.ShortURLStatusExpired):
		return "", &ErrInvalidUpdate{
			Message: "status must be ACTIVE or DISABLED; scheduled and expired are derived from starts_at and expires_at",
		}
	default:
		return "", &ErrInvalidUpdate{Message: "status must be ACTIVE or DISABLED"}
	}
}

// ResolveShortCode maps a short code to the url it should redirect to,
// returning the short url's id so the click can be attributed.
//
// A cache hit serves the redirect without touching the database at all. Every
// cache failure short of a cached "not found" falls through to postgres, so a
// redis outage degrades to plain db lookups rather than breaking redirects.
func (s *UrlService) ResolveShortCode(ctx context.Context, shortCode string) (string, int, error) {
	cached, err := s.cache.Get(ctx, shortCode)
	switch {
	case err == nil:
		return cached.DestinationURL, cached.ShortURLID, nil
	case errors.Is(err, cache.ErrCachedNotFound):
		return "", 0, ErrShortURLNotResolvable
	case !errors.Is(err, cache.ErrCacheMiss):
		log.Printf("short url cache lookup failed for %q: %v\n", shortCode, err)
	}

	return s.resolveFromDB(ctx, shortCode)
}

func (s *UrlService) resolveFromDB(ctx context.Context, shortCode string) (string, int, error) {
	shortURL, err := s.repository.ShortURLs.GetByShortCode(shortCode)
	if errors.Is(err, sql.ErrNoRows) {
		s.cacheNotFound(ctx, shortCode)
		return "", 0, ErrShortURLNotResolvable
	}
	if err != nil {
		return "", 0, err
	}

	if !shortURL.IsResolvable(time.Now()) {
		return "", 0, ErrShortURLNotResolvable
	}

	destination, err := s.repository.DestinationURLs.GetByID(shortURL.DestinationID)
	if errors.Is(err, sql.ErrNoRows) {
		return "", 0, ErrShortURLNotResolvable
	}
	if err != nil {
		return "", 0, err
	}

	s.warmCache(ctx, shortURL, destination.DestinationURL)

	return destination.DestinationURL, shortURL.ID, nil
}

// warmCache caches a short url that is resolvable right now. Scheduled,
// disabled or expired short urls are never cached - serving one of those from
// cache would redirect for a campaign that isn't running. Any existing entry
// (including a stale "not found" sentinel) is cleared in that case.
//
// Caching is best effort: a failure only costs the next request a db lookup,
// so it is logged rather than surfaced.
func (s *UrlService) warmCache(ctx context.Context, shortURL *models.ShortURL, destinationURL string) {
	if !shortURL.IsResolvable(time.Now()) {
		if err := s.cache.Delete(ctx, shortURL.ShortCode); err != nil {
			log.Printf("failed to clear short url cache for %q: %v\n", shortURL.ShortCode, err)
		}
		return
	}

	entry := cache.ResolvedShortURL{
		DestinationURL: destinationURL,
		ShortURLID:     shortURL.ID,
	}

	if err := s.cache.Set(ctx, shortURL.ShortCode, entry, cacheTTL(shortURL.ExpiresAt)); err != nil {
		log.Printf("failed to cache short url %q: %v\n", shortURL.ShortCode, err)
	}
}

func (s *UrlService) cacheNotFound(ctx context.Context, shortCode string) {
	if err := s.cache.SetNotFound(ctx, shortCode); err != nil {
		log.Printf("failed to cache miss for %q: %v\n", shortCode, err)
	}
}

// cacheTTL caps entries at cache.DefaultTTL, shortening it so an entry expires
// exactly when its campaign does. The cap bounds how long a stale redirect can
// be served, since nothing invalidates the cache when a short url changes.
func cacheTTL(expiresAt *time.Time) time.Duration {
	ttl := cache.DefaultTTL

	if expiresAt != nil {
		if until := time.Until(*expiresAt); until < ttl {
			ttl = until
		}
	}

	return ttl
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
