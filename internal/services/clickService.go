package services

import (
	"context"
	"time"

	"github.com/dhruv15803/url-shortener-service/internal/geoip"
	"github.com/dhruv15803/url-shortener-service/internal/models"
	"github.com/dhruv15803/url-shortener-service/internal/queue"
	"github.com/dhruv15803/url-shortener-service/internal/repositories"
	"github.com/dhruv15803/url-shortener-service/internal/useragent"
)

// RecordClickInput carries the raw request details the handler pulled off the
// wire, so the service layer never has to know about *http.Request.
type RecordClickInput struct {
	ShortURLID int
	IP         string
	UserAgent  string
	Referrer   string
}

type ClickService struct {
	repository *repositories.Repository
	queue      *queue.ClickQueue
	geoip      *geoip.Reader
}

func NewClickService(repository *repositories.Repository, clickQueue *queue.ClickQueue, geoipReader *geoip.Reader) *ClickService {
	return &ClickService{
		repository: repository,
		queue:      clickQueue,
		geoip:      geoipReader,
	}
}

// RecordClick enriches the raw click details and hands the event to the queue.
// The api calls this off the request path; the db write happens in the worker.
func (s *ClickService) RecordClick(ctx context.Context, input RecordClickInput) error {
	browser, os, device := useragent.Parse(input.UserAgent)
	country, city, region := s.geoip.Lookup(input.IP)

	event := models.ClickEvent{
		ShortURLID: input.ShortURLID,
		ClickedAt:  time.Now(),
		IP:         nilIfEmpty(input.IP),
		UserAgent:  nilIfEmpty(input.UserAgent),
		Referrer:   nilIfEmpty(input.Referrer),
		Country:    country,
		City:       city,
		Region:     region,
		Device:     device,
		Browser:    browser,
		OS:         os,
	}

	return s.queue.Publish(ctx, event)
}

// PersistClick is the worker's queue handler - the only place clicks are
// written to the database.
func (s *ClickService) PersistClick(event models.ClickEvent) error {
	return s.repository.Clicks.Create(event)
}

// nilIfEmpty keeps empty values out of the database as nulls rather than
// storing meaningless empty strings.
func nilIfEmpty(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}
