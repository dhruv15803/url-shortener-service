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
	"github.com/dhruv15803/url-shortener-service/internal/optional"
	"github.com/dhruv15803/url-shortener-service/internal/repositories"
	"github.com/dhruv15803/url-shortener-service/internal/services"
	"github.com/go-chi/chi/v5"
)

const (
	defaultCampaignLimit = 50
	maxCampaignLimit     = 100

	// maxSearchLength matches the width of short_urls.name - the longest thing
	// a search term could usefully match in full.
	maxSearchLength = 255
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

type updateCampaignRequest struct {
	Name      optional.Optional[string]    `json:"name"`
	Status    optional.Optional[string]    `json:"status"`
	StartsAt  optional.Optional[time.Time] `json:"starts_at"`
	ExpiresAt optional.Optional[time.Time] `json:"expires_at"`
}

// campaignResponse is the shape every campaign-facing endpoint returns, so the
// frontend deals with one consistent object. Status here is always derived.
type campaignResponse struct {
	ID             int        `json:"id"`
	Code           string     `json:"code"`
	ShortURL       string     `json:"short_url"`
	CampaignName   *string    `json:"campaign_name"`
	Status         string     `json:"status"`
	DestinationURL string     `json:"destination_url"`
	StartsAt       *time.Time `json:"starts_at"`
	ExpiresAt      *time.Time `json:"expires_at"`
	CreatedAt      time.Time  `json:"created_at"`
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

// RegisterRoutes mounts the url routes on their own sub-routers so the auth
// middleware applies to them only, and not to the routes other handlers
// register on the shared /api router.
//
// /urls is keyed by short code and /destinations by destination id, so the two
// identifier types never share a path position.
func (h *UrlHandler) RegisterRoutes(r chi.Router) {
	r.Route("/urls", func(r chi.Router) {
		r.Use(middleware.Auth(h.jwtSecret))

		r.Get("/", h.ListCampaigns)
		r.Post("/", h.CreateOrGetDestination)
		r.Get("/{shortCode}", h.GetCampaign)
		r.Put("/{shortCode}", h.UpdateCampaign)
		r.Get("/{shortCode}/clicks", h.GetCampaignClicks)
	})

	r.Route("/destinations", func(r chi.Router) {
		r.Use(middleware.Auth(h.jwtSecret))

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

	result, err := h.service.Urls.FindOrCreateDestination(r.Context(), userID, services.CreateDestinationInput{
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

	shortURL, err := h.service.Urls.CreateShortURLForDestination(r.Context(), userID, destinationID, services.CreateShortURLInput{
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

func (h *UrlHandler) ListCampaigns(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		httpresponse.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	limit := intQueryParam(r, "limit", defaultCampaignLimit)
	if limit < 1 {
		limit = defaultCampaignLimit
	}
	if limit > maxCampaignLimit {
		limit = maxCampaignLimit
	}

	offset := intQueryParam(r, "offset", 0)
	if offset < 0 {
		offset = 0
	}

	filter := repositories.CampaignFilter{Search: searchQueryParam(r, "search")}

	// Unlike limit and offset, an unrecognised status is rejected rather than
	// ignored: silently dropping it would answer with the full unfiltered list,
	// which reads as the filter being broken.
	if raw := strings.TrimSpace(r.URL.Query().Get("status")); raw != "" {
		status, ok := models.ParseShortURLStatus(raw)
		if !ok {
			httpresponse.WriteError(w, http.StatusBadRequest, "status must be one of active, scheduled, expired, disabled")
			return
		}
		filter.Status = status
	}

	campaigns, total, err := h.service.Urls.ListCampaigns(userID, filter, limit, offset)
	if err != nil {
		httpresponse.WriteError(w, http.StatusInternalServerError, "failed to list campaigns")
		return
	}

	responses := make([]campaignResponse, 0, len(campaigns))
	for _, campaign := range campaigns {
		responses = append(responses, h.toCampaignResponse(campaign))
	}

	httpresponse.WriteJSON(w, http.StatusOK, map[string]any{
		"campaigns": responses,
		"total":     total,
		"limit":     limit,
		"offset":    offset,
	})
}

func (h *UrlHandler) GetCampaign(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		httpresponse.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	campaign, err := h.service.Urls.GetCampaign(userID, chi.URLParam(r, "shortCode"))
	if errors.Is(err, services.ErrShortURLNotFound) {
		httpresponse.WriteError(w, http.StatusNotFound, "short url not found")
		return
	}
	if err != nil {
		httpresponse.WriteError(w, http.StatusInternalServerError, "failed to fetch campaign")
		return
	}

	httpresponse.WriteJSON(w, http.StatusOK, map[string]any{"campaign": h.toCampaignResponse(campaign)})
}

func (h *UrlHandler) UpdateCampaign(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		httpresponse.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req updateCampaignRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpresponse.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	campaign, err := h.service.Urls.UpdateShortURL(r.Context(), userID, chi.URLParam(r, "shortCode"), services.UpdateShortURLInput{
		Name:      req.Name,
		Status:    req.Status,
		StartsAt:  req.StartsAt,
		ExpiresAt: req.ExpiresAt,
	})

	var invalid *services.ErrInvalidUpdate
	switch {
	case errors.Is(err, services.ErrShortURLNotFound):
		httpresponse.WriteError(w, http.StatusNotFound, "short url not found")
		return
	case errors.As(err, &invalid):
		httpresponse.WriteError(w, http.StatusBadRequest, invalid.Message)
		return
	case err != nil:
		httpresponse.WriteError(w, http.StatusInternalServerError, "failed to update campaign")
		return
	}

	httpresponse.WriteJSON(w, http.StatusOK, map[string]any{"campaign": h.toCampaignResponse(campaign)})
}

func (h *UrlHandler) toCampaignResponse(campaign *services.CampaignView) campaignResponse {
	return campaignResponse{
		ID:             campaign.ShortURL.ID,
		Code:           campaign.ShortURL.ShortCode,
		ShortURL:       h.shortURLBaseURL + "/" + campaign.ShortURL.ShortCode,
		CampaignName:   campaign.ShortURL.Name,
		Status:         string(campaign.Status),
		DestinationURL: campaign.DestinationURL,
		StartsAt:       campaign.ShortURL.StartsAt,
		ExpiresAt:      campaign.ShortURL.ExpiresAt,
		CreatedAt:      campaign.ShortURL.CreatedAt,
	}
}

// searchQueryParam reads a free-text filter. Whitespace-only is treated as
// absent, and the value is capped so an oversized term can't be pushed into
// the query. The cap counts runes, since slicing bytes could split one.
func searchQueryParam(r *http.Request, name string) string {
	value := strings.TrimSpace(r.URL.Query().Get(name))
	if runes := []rune(value); len(runes) > maxSearchLength {
		value = string(runes[:maxSearchLength])
	}

	return value
}

func intQueryParam(r *http.Request, name string, fallback int) int {
	raw := r.URL.Query().Get(name)
	if raw == "" {
		return fallback
	}

	value, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}

	return value
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
