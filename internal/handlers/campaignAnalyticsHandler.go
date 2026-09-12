package handlers

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/dhruv15803/url-shortener-service/internal/httpresponse"
	"github.com/dhruv15803/url-shortener-service/internal/middleware"
	"github.com/dhruv15803/url-shortener-service/internal/repositories"
	"github.com/dhruv15803/url-shortener-service/internal/services"
	"github.com/go-chi/chi/v5"
)

// GetCampaignClicks serves GET /api/urls/{shortCode}/clicks.
//
// It hangs off UrlHandler rather than its own handler struct because chi can't
// mount a second router at /urls/{shortCode}/clicks while /urls is already
// mounted on the same parent - the patterns overlap.
func (h *UrlHandler) GetCampaignClicks(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		httpresponse.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	query := services.ClickAnalyticsQuery{
		ShortCode: chi.URLParam(r, "shortCode"),
		Limit:     intQueryParam(r, "limit", services.DefaultAnalyticsLimit),
		Offset:    intQueryParam(r, "offset", 0),
	}

	if raw := strings.TrimSpace(r.URL.Query().Get("group_by")); raw != "" {
		dimension, valid := repositories.ParseClickDimension(strings.ToLower(raw))
		if !valid {
			httpresponse.WriteError(w, http.StatusBadRequest,
				"group_by must be one of: "+strings.Join(repositories.AllClickDimensions(), ", "))
			return
		}
		query.GroupBy = &dimension
	}

	startTS, err := parseTimestampParam(r, "start_ts")
	if err != nil {
		httpresponse.WriteError(w, http.StatusBadRequest, "start_ts must be an RFC3339 timestamp")
		return
	}
	query.StartTS = startTS

	endTS, err := parseTimestampParam(r, "end_ts")
	if err != nil {
		httpresponse.WriteError(w, http.StatusBadRequest, "end_ts must be an RFC3339 timestamp")
		return
	}
	query.EndTS = endTS

	result, err := h.service.Analytics.ClickAnalytics(r.Context(), userID, query)

	var invalid *services.ErrInvalidUpdate
	switch {
	case errors.Is(err, services.ErrShortURLNotFound):
		httpresponse.WriteError(w, http.StatusNotFound, "short url not found")
		return
	case errors.As(err, &invalid):
		httpresponse.WriteError(w, http.StatusBadRequest, invalid.Message)
		return
	case err != nil:
		httpresponse.WriteError(w, http.StatusInternalServerError, "failed to load click analytics")
		return
	}

	httpresponse.WriteJSON(w, http.StatusOK, result)
}

func parseTimestampParam(r *http.Request, name string) (*time.Time, error) {
	raw := strings.TrimSpace(r.URL.Query().Get(name))
	if raw == "" {
		return nil, nil
	}

	parsed, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return nil, err
	}

	return &parsed, nil
}
