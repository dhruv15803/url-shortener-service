package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	analyticsKeyPrefix = "analytics:"

	// Short by design: a dashboard refresh or a dimension switch shouldn't
	// re-scan the clicks table, but the numbers should still feel live.
	AnalyticsTTL = 60 * time.Second
)

type AnalyticsCache struct {
	client *redis.Client
}

func NewAnalyticsCache(client *redis.Client) *AnalyticsCache {
	return &AnalyticsCache{
		client: client,
	}
}

// Key identifies one analytics answer. The range must already be normalised
// and rounded by the caller - keying on a raw "now" would miss on every
// request as the clock moves.
func AnalyticsKey(shortURLID int, groupBy string, start time.Time, end time.Time) string {
	if groupBy == "" {
		groupBy = "overview"
	}
	return fmt.Sprintf("%s%d:%s:%d:%d", analyticsKeyPrefix, shortURLID, groupBy, start.Unix(), end.Unix())
}

// Get decodes a cached payload into out. ErrCacheMiss means "ask postgres";
// every other error is also safe to treat that way.
func (c *AnalyticsCache) Get(ctx context.Context, key string, out any) error {
	ctx, cancel := context.WithTimeout(ctx, OpTimeout)
	defer cancel()

	payload, err := c.client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return ErrCacheMiss
	}
	if err != nil {
		return err
	}

	if err := json.Unmarshal([]byte(payload), out); err != nil {
		return ErrCacheMiss
	}

	return nil
}

func (c *AnalyticsCache) Set(ctx context.Context, key string, value any) error {
	payload, err := json.Marshal(value)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(ctx, OpTimeout)
	defer cancel()

	return c.client.Set(ctx, key, payload, AnalyticsTTL).Err()
}
