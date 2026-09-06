package handlers

import (
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/dhruv15803/url-shortener-service/internal/httpresponse"
	"github.com/dhruv15803/url-shortener-service/internal/middleware"
	"github.com/dhruv15803/url-shortener-service/internal/services"
	"github.com/go-chi/chi/v5"
)

const (
	oauthStateCookieName = "oauth_state"
	oauthStateTTL        = 10 * time.Minute
	sessionTTL           = 24 * time.Hour
)

type AuthHandler struct {
	service     *services.Service
	jwtSecret   string
	frontendURL string
}

func NewAuthHandler(service *services.Service, jwtSecret string, frontendURL string) *AuthHandler {
	return &AuthHandler{
		service:     service,
		jwtSecret:   jwtSecret,
		frontendURL: frontendURL,
	}
}

func (h *AuthHandler) RegisterRoutes(r chi.Router) {
	r.Get("/auth/google/login", h.GoogleLogin)
	r.Get("/auth/callback", h.GoogleCallback)

	// Logout is deliberately unauthenticated so an expired or malformed
	// session cookie can still be cleared.
	r.Post("/auth/logout", h.Logout)

	r.Group(func(r chi.Router) {
		r.Use(middleware.Auth(h.jwtSecret))
		r.Get("/auth/me", h.Me)
	})
}

// Me lets the frontend bootstrap auth state on load: the session cookie is
// httpOnly, so the browser can't inspect it directly.
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		httpresponse.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	user, err := h.service.Users.GetUserByID(userID)
	if errors.Is(err, services.ErrUserNotFound) {
		httpresponse.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	if err != nil {
		httpresponse.WriteError(w, http.StatusInternalServerError, "failed to load user")
		return
	}

	httpresponse.WriteJSON(w, http.StatusOK, map[string]any{"user": user})
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     middleware.SessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})

	w.WriteHeader(http.StatusNoContent)
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

	// This endpoint is reached by a browser redirect from Google, so every
	// outcome has to land the user back in the app rather than on a json body.
	if errParam := r.URL.Query().Get("error"); errParam != "" {
		http.SetCookie(w, clearStateCookie)
		h.redirectToFrontend(w, r, "login_cancelled")
		return
	}

	stateCookie, err := r.Cookie(oauthStateCookieName)
	if err != nil || stateCookie.Value == "" || stateCookie.Value != r.URL.Query().Get("state") {
		http.SetCookie(w, clearStateCookie)
		h.redirectToFrontend(w, r, "invalid_state")
		return
	}
	http.SetCookie(w, clearStateCookie)

	code := r.URL.Query().Get("code")
	if code == "" {
		h.redirectToFrontend(w, r, "missing_code")
		return
	}

	_, jwtToken, err := h.service.Auth.HandleGoogleCallback(r.Context(), code)
	if err != nil {
		h.redirectToFrontend(w, r, "login_failed")
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     middleware.SessionCookieName,
		Value:    jwtToken,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(sessionTTL),
	})

	h.redirectToFrontend(w, r, "")
}

// redirectToFrontend sends the browser back to the app, tagging the login
// route with an error code when sign-in didn't succeed.
func (h *AuthHandler) redirectToFrontend(w http.ResponseWriter, r *http.Request, errCode string) {
	target := h.frontendURL

	if errCode != "" {
		target = strings.TrimRight(h.frontendURL, "/") + "/login?error=" + url.QueryEscape(errCode)
	}

	http.Redirect(w, r, target, http.StatusFound)
}
