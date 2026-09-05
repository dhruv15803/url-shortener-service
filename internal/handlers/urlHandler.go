package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/dhruv15803/url-shortener-service/internal/httpresponse"
	"github.com/dhruv15803/url-shortener-service/internal/middleware"
	"github.com/dhruv15803/url-shortener-service/internal/models"
	"github.com/dhruv15803/url-shortener-service/internal/services"
	"github.com/go-chi/chi/v5"
)

type createDestinationRequest struct {
	DestinationURL string     `json:"destination_url"`
	CampaignName   *string    `json:"campaign_name"`
	StartsAt       *time.Time `json:"starts_at"`
	ExpiresAt      *time.Time `json:"expires_at"`
}

type createShortURLRequest struct {
	CampaignName *string    `json:"campaign_name"`
	StartsAt     *time.Time `json:"starts_at"`
	ExpiresAt    *time.Time `json:"expires_at"`
}

type destinationURLResponse struct {
	ID  int    `json:"id"`
	URL string `json:"url"`
}

type shortURLResponse struct {
	ID           int     `json:"id"`
	Code         string  `json:"code"`
	URL          string  `json:"url"`
	CampaignName *string `json:"campaign_name,omitempty"`
	Status       string  `json:"status"`
}

type UrlHandler struct {
	service         *services.Service
	jwtSecret       string
	shortURLBaseURL string
}

func NewUrlHandler(service *services.Service, jwtSecret string, shortURLBaseURL string) *UrlHandler {
	return &UrlHandler{
		service:         service,
		jwtSecret:       jwtSecret,
		shortURLBaseURL: shortURLBaseURL,
	}
}

// RegisterRoutes mounts the url routes on their own sub-router so the auth
// middleware applies to them only, and not to the routes other handlers
// register on the shared /api router.
func (h *UrlHandler) RegisterRoutes(r chi.Router) {
	r.Route("/urls", func(r chi.Router) {
		r.Use(middleware.Auth(h.jwtSecret))

		r.Post("/", h.CreateOrGetDestination)
		r.Post("/{destinationId}/short-urls", h.CreateShortURLForDestination)
	})
}

func (h *UrlHandler) CreateOrGetDestination(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		httpresponse.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req createDestinationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpresponse.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	destinationURL := strings.TrimSpace(req.DestinationURL)
	if destinationURL == "" {
		httpresponse.WriteError(w, http.StatusBadRequest, "destination_url is required")
		return
	}

	result, err := h.service.Urls.FindOrCreateDestination(userID, services.CreateDestinationInput{
		DestinationURL: destinationURL,
		CampaignName:   req.CampaignName,
		StartsAt:       req.StartsAt,
		ExpiresAt:      req.ExpiresAt,
	})
	if err != nil {
		httpresponse.WriteError(w, http.StatusInternalServerError, "failed to create or fetch destination url")
		return
	}

	destination := destinationURLResponse{
		ID:  result.Destination.ID,
		URL: result.Destination.DestinationURL,
	}

	if !result.Created {
		shortURLs := make([]shortURLResponse, 0, len(result.ExistingShortURLs))
		for _, shortURL := range result.ExistingShortURLs {
			shortURLs = append(shortURLs, h.toShortURLResponse(shortURL))
		}

		httpresponse.WriteJSON(w, http.StatusOK, map[string]any{
			"destination_url": destination,
			"short_urls":      shortURLs,
		})
		return
	}

	httpresponse.WriteJSON(w, http.StatusCreated, map[string]any{
		"destination_url": destination,
		"short_url":       h.toShortURLResponse(result.NewShortURL),
	})
}

func (h *UrlHandler) CreateShortURLForDestination(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		httpresponse.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	destinationID, err := strconv.Atoi(chi.URLParam(r, "destinationId"))
	if err != nil {
		httpresponse.WriteError(w, http.StatusBadRequest, "invalid destination id")
		return
	}

	var req createShortURLRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpresponse.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	shortURL, err := h.service.Urls.CreateShortURLForDestination(userID, destinationID, services.CreateShortURLInput{
		CampaignName: req.CampaignName,
		StartsAt:     req.StartsAt,
		ExpiresAt:    req.ExpiresAt,
	})
	if errors.Is(err, services.ErrDestinationNotFound) {
		httpresponse.WriteError(w, http.StatusNotFound, "destination not found")
		return
	}
	if err != nil {
		httpresponse.WriteError(w, http.StatusInternalServerError, "failed to create short url")
		return
	}

	httpresponse.WriteJSON(w, http.StatusCreated, map[string]any{
		"short_url": h.toShortURLResponse(shortURL),
	})
}

func (h *UrlHandler) toShortURLResponse(shortURL *models.ShortURL) shortURLResponse {
	return shortURLResponse{
		ID:           shortURL.ID,
		Code:         shortURL.ShortCode,
		URL:          h.shortURLBaseURL + "/" + shortURL.ShortCode,
		CampaignName: shortURL.Name,
		Status:       string(shortURL.Status),
	}
}
