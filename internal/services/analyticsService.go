package services

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/dhruv15803/url-shortener-service/internal/cache"
	"github.com/dhruv15803/url-shortener-service/internal/repositories"
)

const (
	defaultAnalyticsWindow = 7 * 24 * time.Hour
	maxAnalyticsWindow     = 90 * 24 * time.Hour

	// Below this the series is hourly; above it, daily. Keeps the point count
	// bounded either way (~48 hourly, ~90 daily).
	hourlyBucketThreshold = 48 * time.Hour

	DefaultAnalyticsLimit = 20
	MaxAnalyticsLimit     = 100
)

type ClickAnalyticsQuery struct {
	ShortCode string
	GroupBy   *repositories.ClickDimension // nil => overview
	StartTS   *time.Time
	EndTS     *time.Time
	Limit     int
	Offset    int
}

type TimeRange struct {
	StartTS time.Time `json:"start_ts"`
	EndTS   time.Time `json:"end_ts"`
}

type DimensionSlice struct {
	Value      string  `json:"value"`
	Clicks     int     `json:"clicks"`
	Percentage float64 `json:"percentage"`
}

type SeriesPoint struct {
	Bucket time.Time `json:"bucket"`
	Clicks int       `json:"clicks"`
}

type ClickAnalyticsResult struct {
	Range       TimeRange        `json:"range"`
	TotalClicks int              `json:"total_clicks"`
	GroupBy     string           `json:"group_by,omitempty"`
	Data        []DimensionSlice `json:"data,omitempty"`
	Series      []SeriesPoint    `json:"series,omitempty"`

	// Paging metadata, breakdowns only. TotalGroups is the number of distinct
	// values - the row count, not the click count - which is what the client
	// needs to work out how many pages there are.
	//
	// Pointers so a breakdown always reports all three (offset 0 and
	// total_groups 0 included) while an overview reports none. With plain ints
	// and omitempty, page one would silently drop "offset": 0.
	TotalGroups *int `json:"total_groups,omitempty"`
	Limit       *int `json:"limit,omitempty"`
	Offset      *int `json:"offset,omitempty"`
}

type AnalyticsService struct {
	repository *repositories.Repository
	urls       *UrlService
	cache      *cache.AnalyticsCache
}

func NewAnalyticsService(repository *repositories.Repository, urls *UrlService, analyticsCache *cache.AnalyticsCache) *AnalyticsService {
	return &AnalyticsService{
		repository: repository,
		urls:       urls,
		cache:      analyticsCache,
	}
}

// ClickAnalytics returns either an overview (total + trend) or a single
// dimension breakdown, for a campaign the caller owns.
func (s *AnalyticsService) ClickAnalytics(ctx context.Context, userID int, query ClickAnalyticsQuery) (*ClickAnalyticsResult, error) {
	// Same not-found-or-not-owned rule as the rest of the campaign endpoints,
	// so analytics can't be used to probe other users' short codes.
	shortURL, _, err := s.urls.ownedShortURL(userID, query.ShortCode)
	if err != nil {
		return nil, err
	}

	timeRange, err := resolveRange(query.StartTS, query.EndTS)
	if err != nil {
		return nil, err
	}

	bucket := bucketFor(timeRange)

	// Round to the bucket so a rolling "now" doesn't produce a new cache key
	// on every request.
	cacheRange := TimeRange{
		StartTS: truncateTo(timeRange.StartTS, bucket),
		EndTS:   truncateTo(timeRange.EndTS, bucket),
	}

	groupBy := ""
	if query.GroupBy != nil {
		groupBy = string(*query.GroupBy)
	}

	limit := resolveLimit(query.Limit)
	offset := resolveOffset(query.Offset)

	key := cache.AnalyticsKey(shortURL.ID, groupBy, cacheRange.StartTS, cacheRange.EndTS, limit, offset)

	var cached ClickAnalyticsResult
	if err := s.cache.Get(ctx, key, &cached); err == nil {
		return &cached, nil
	} else if !errors.Is(err, cache.ErrCacheMiss) {
		log.Printf("analytics cache lookup failed for %q: %v\n", key, err)
	}

	total, err := s.repository.Clicks.CountByShortURL(shortURL.ID, timeRange.StartTS, timeRange.EndTS)
	if err != nil {
		return nil, err
	}

	result := &ClickAnalyticsResult{Range: timeRange, TotalClicks: total}
	// total = 500 , for a dimesion = "country"
	// "india" , "usa" , "uk" , null -> 500 / 20  = 25 pages

	if query.GroupBy == nil {
		points, err := s.repository.Clicks.SeriesByBucket(shortURL.ID, bucket, timeRange.StartTS, timeRange.EndTS)
		if err != nil {
			return nil, err
		}

		result.Series = make([]SeriesPoint, 0, len(points))
		for _, point := range points {
			result.Series = append(result.Series, SeriesPoint{Bucket: point.Bucket, Clicks: point.Clicks})
		}
	} else {
		counts, err := s.repository.Clicks.GroupByDimension(shortURL.ID, *query.GroupBy, timeRange.StartTS, timeRange.EndTS, limit, offset)
		if err != nil {
			return nil, err
		}

		totalGroups, err := s.repository.Clicks.CountDimensionGroups(shortURL.ID, *query.GroupBy, timeRange.StartTS, timeRange.EndTS)
		if err != nil {
			return nil, err
		}

		result.GroupBy = groupBy
		result.TotalGroups = &totalGroups
		result.Limit = &limit
		result.Offset = &offset
		result.Data = make([]DimensionSlice, 0, len(counts))
		for _, count := range counts {
			result.Data = append(result.Data, DimensionSlice{
				Value:      count.Value,
				Clicks:     count.Clicks,
				Percentage: percentageOf(count.Clicks, total),
			})
		}
	}

	if err := s.cache.Set(ctx, key, result); err != nil {
		log.Printf("failed to cache analytics for %q: %v\n", key, err)
	}

	return result, nil
}

// resolveRange applies the defaults and refuses windows that would be
// expensive to scan, rather than silently truncating them.
func resolveRange(start *time.Time, end *time.Time) (TimeRange, error) {
	endTS := time.Now()
	if end != nil {
		endTS = *end
	}

	startTS := endTS.Add(-defaultAnalyticsWindow)
	if start != nil {
		startTS = *start
	}

	if !startTS.Before(endTS) {
		return TimeRange{}, &ErrInvalidUpdate{Message: "start_ts must be before end_ts"}
	}

	if endTS.Sub(startTS) > maxAnalyticsWindow {
		return TimeRange{}, &ErrInvalidUpdate{
			Message: fmt.Sprintf("time range must be %d days or less", int(maxAnalyticsWindow.Hours()/24)),
		}
	}

	return TimeRange{StartTS: startTS, EndTS: endTS}, nil
}

func bucketFor(timeRange TimeRange) repositories.BucketSize {
	if timeRange.EndTS.Sub(timeRange.StartTS) <= hourlyBucketThreshold {
		return repositories.BucketHour
	}
	return repositories.BucketDay
}

func truncateTo(at time.Time, bucket repositories.BucketSize) time.Time {
	if bucket == repositories.BucketHour {
		return at.Truncate(time.Hour)
	}
	return at.Truncate(24 * time.Hour)
}

func resolveLimit(limit int) int {
	if limit <= 0 {
		return DefaultAnalyticsLimit
	}
	if limit > MaxAnalyticsLimit {
		return MaxAnalyticsLimit
	}
	return limit
}

func resolveOffset(offset int) int {
	if offset < 0 {
		return 0
	}
	return offset
}

func percentageOf(clicks int, total int) float64 {
	if total == 0 {
		return 0
	}
	// Two decimals is plenty for a dashboard and keeps the json tidy.
	return float64(int((float64(clicks)/float64(total))*10000+0.5)) / 100
}
