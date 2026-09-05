package queue

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"time"

	"github.com/dhruv15803/url-shortener-service/internal/models"
	"github.com/redis/go-redis/v9"
)

const (
	ClickQueueKey      = "clicks:queue"
	ClickProcessingKey = "clicks:processing"
)

type ClickQueue struct {
	client *redis.Client
}

func NewClickQueue(client *redis.Client) *ClickQueue {
	return &ClickQueue{
		client: client,
	}
}

// Publish serializes the event and pushes it onto the queue. Callers on the
// request path must treat a failure here as non-fatal - a dropped analytics
// event must never break a redirect.
func (q *ClickQueue) Publish(ctx context.Context, event models.ClickEvent) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return q.client.LPush(ctx, ClickQueueKey, payload).Err()
}

// Consume blocks, draining the queue until ctx is cancelled.
//
// It uses the reliable-queue pattern: an entry is atomically moved to a
// processing list while it is being handled, and only removed from there once
// the handler succeeds. If the worker dies mid-write the entry survives in
// clicks:processing instead of being lost.
func (q *ClickQueue) Consume(ctx context.Context, handle func(models.ClickEvent) error) {
	for {
		if ctx.Err() != nil {
			return
		}

		payload, err := q.client.BRPopLPush(ctx, ClickQueueKey, ClickProcessingKey, 0).Result()
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return
			}
			// redis unavailable - back off briefly rather than spinning hot.
			log.Printf("click queue: pop failed: %v\n", err)
			time.Sleep(time.Second)
			continue
		}

		var event models.ClickEvent
		if err := json.Unmarshal([]byte(payload), &event); err != nil {
			// Undecodable payload would block forever if left in place.
			log.Printf("click queue: discarding malformed payload: %v\n", err)
			q.client.LRem(ctx, ClickProcessingKey, 1, payload)
			continue
		}

		if err := handle(event); err != nil {
			// Left in clicks:processing on purpose, for inspection.
			log.Printf("click queue: handler failed for short_url_id=%d: %v\n", event.ShortURLID, err)
			continue
		}

		if err := q.client.LRem(ctx, ClickProcessingKey, 1, payload).Err(); err != nil {
			log.Printf("click queue: failed to clear processed entry: %v\n", err)
		}
	}
}
