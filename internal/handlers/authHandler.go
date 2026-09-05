package handlers

import (
	"net/http"
	"time"

	"github.com/dhruv15803/url-shortener-service/internal/httpresponse"
	"github.com/dhruv15803/url-shortener-service/internal/services"
	"github.com/go-chi/chi/v5"
)

const (
	oauthStateCookieName = "oauth_state"
	sessionCookieName    = "session"
	oauthStateTTL        = 10 * time.Minute
	sessionTTL           = 24 * time.Hour
)

type AuthHandler struct {
	service *services.Service
}

func NewAuthHandler(service *services.Service) *AuthHandler {
	return &AuthHandler{
		service: service,
	}
}

func (h *AuthHandler) RegisterRoutes(r chi.Router) {
	r.Get("/auth/google/login", h.GoogleLogin)
	r.Get("/auth/callback", h.GoogleCallback)
}

func (h *AuthHandler) GoogleLogin(w http.ResponseWriter, r *http.Request) {
	state, err := h.service.Auth.GenerateState()
	if err != nil {
		httpresponse.WriteError(w, http.StatusInternalServerError, "failed to start google login")
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     oauthStateCookieName,
		Value:    state,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(oauthStateTTL),
	})

	http.Redirect(w, r, h.service.Auth.GoogleAuthCodeURL(state), http.StatusTemporaryRedirect)
}

func (h *AuthHandler) GoogleCallback(w http.ResponseWriter, r *http.Request) {
	clearStateCookie := &http.Cookie{
		Name:     oauthStateCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	}

	if errParam := r.URL.Query().Get("error"); errParam != "" {
		http.SetCookie(w, clearStateCookie)
		httpresponse.WriteError(w, http.StatusBadRequest, "google login was cancelled or denied")
		return
	}

	stateCookie, err := r.Cookie(oauthStateCookieName)
	if err != nil || stateCookie.Value == "" || stateCookie.Value != r.URL.Query().Get("state") {
		http.SetCookie(w, clearStateCookie)
		httpresponse.WriteError(w, http.StatusBadRequest, "invalid oauth state")
		return
	}
	http.SetCookie(w, clearStateCookie)

	code := r.URL.Query().Get("code")
	if code == "" {
		httpresponse.WriteError(w, http.StatusBadRequest, "missing code")
		return
	}

	user, jwtToken, err := h.service.Auth.HandleGoogleCallback(r.Context(), code)
	if err != nil {
		httpresponse.WriteError(w, http.StatusInternalServerError, "google login failed")
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    jwtToken,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(sessionTTL),
	})

	httpresponse.WriteJSON(w, http.StatusOK, map[string]any{"user": user})
}
