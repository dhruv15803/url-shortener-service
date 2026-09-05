package cache

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	shortURLKeyPrefix = "shorturl:"

	// notFoundValue marks a code we've already looked up and failed to
	// resolve, so a burst of requests for unknown codes can't fall through
	// to postgres every time.
	notFoundValue = "__not_found__"

	// DefaultTTL caps how long any entry can live. It bounds how stale a
	// cached redirect can get, since there is no invalidation path yet.
	DefaultTTL = 15 * time.Minute

	NotFoundTTL = 60 * time.Second

	// OpTimeout bounds every cache call. The cache sits on the redirect hot
	// path, so when redis is unreachable a request must give up almost
	// immediately and fall through to postgres. Without this, callers wait
	// out the redis client's dial timeout and retries on every request,
	// turning a redis outage into seconds-long redirects.
	OpTimeout = 150 * time.Millisecond
)

var (
	ErrCacheMiss      = errors.New("cache miss")
	ErrCachedNotFound = errors.New("short code cached as not found")
)

// ResolvedShortURL is everything the redirect hot path needs: where to send
// the visitor, and which short url to attribute the click to. Caching both is
// what lets a redirect be served without touching the database.
type ResolvedShortURL struct {
	DestinationURL string `json:"destination_url"`
	ShortURLID     int    `json:"short_url_id"`
}

type ShortURLCache struct {
	client *redis.Client
}

func NewShortURLCache(client *redis.Client) *ShortURLCache {
	return &ShortURLCache{
		client: client,
	}
}

// Get returns the cached entry, ErrCacheMiss when nothing is cached,
// ErrCachedNotFound when the code is known to be unresolvable, or the
// underlying redis error. Callers treat every error except ErrCachedNotFound
// as "go ask the database".
func (c *ShortURLCache) Get(ctx context.Context, shortCode string) (*ResolvedShortURL, error) {
	ctx, cancel := context.WithTimeout(ctx, OpTimeout)
	defer cancel()

	value, err := c.client.Get(ctx, key(shortCode)).Result()
	if errors.Is(err, redis.Nil) {
		return nil, ErrCacheMiss
	}
	if err != nil {
		return nil, err
	}

	if value == notFoundValue {
		return nil, ErrCachedNotFound
	}

	var entry ResolvedShortURL
	if err := json.Unmarshal([]byte(value), &entry); err != nil {
		// A corrupt entry shouldn't wedge the code forever.
		return nil, ErrCacheMiss
	}

	return &entry, nil
}

func (c *ShortURLCache) Set(ctx context.Context, shortCode string, entry ResolvedShortURL, ttl time.Duration) error {
	if ttl <= 0 {
		return nil
	}

	payload, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(ctx, OpTimeout)
	defer cancel()

	return c.client.Set(ctx, key(shortCode), payload, ttl).Err()
}

func (c *ShortURLCache) SetNotFound(ctx context.Context, shortCode string) error {
	ctx, cancel := context.WithTimeout(ctx, OpTimeout)
	defer cancel()

	return c.client.Set(ctx, key(shortCode), notFoundValue, NotFoundTTL).Err()
}

func (c *ShortURLCache) Delete(ctx context.Context, shortCode string) error {
	ctx, cancel := context.WithTimeout(ctx, OpTimeout)
	defer cancel()

	return c.client.Del(ctx, key(shortCode)).Err()
}

func key(shortCode string) string {
	return shortURLKeyPrefix + shortCode
}
