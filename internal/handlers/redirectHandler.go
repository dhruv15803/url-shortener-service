package handlers

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/dhruv15803/url-shortener-service/internal/httpresponse"
	"github.com/dhruv15803/url-shortener-service/internal/services"
	"github.com/go-chi/chi/v5"
)

const recordClickTimeout = 2 * time.Second

type RedirectHandler struct {
	service *services.Service
}

func NewRedirectHandler(service *services.Service) *RedirectHandler {
	return &RedirectHandler{
		service: service,
	}
}

func (h *RedirectHandler) RegisterRoutes(r chi.Router) {
	r.Get("/{shortCode}", h.Redirect)
}

func (h *RedirectHandler) Redirect(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	shortCode := chi.URLParam(r, "shortCode")

	destinationURL, shortURLID, err := h.service.Urls.ResolveShortCode(r.Context(), shortCode)
	if errors.Is(err, services.ErrShortURLNotResolvable) {
		httpresponse.WriteError(w, http.StatusNotFound, "short url not found")
		return
	}
	if err != nil {
		httpresponse.WriteError(w, http.StatusInternalServerError, "failed to resolve short url")
		return
	}

	h.recordClick(services.RecordClickInput{
		ShortURLID: shortURLID,
		IP:         clientIP(r),
		UserAgent:  r.UserAgent(),
		Referrer:   r.Referer(),
	})

	responseTime := time.Since(start)
	log.Printf("redirection response time ms :- %v\n", responseTime.Milliseconds())
	http.Redirect(w, r, destinationURL, http.StatusFound)
}

// recordClick fires the analytics event without blocking the redirect. It runs
// on its own context because r.Context() is cancelled the moment the response
// is written, and a failure here is logged rather than surfaced to the visitor.
func (h *RedirectHandler) recordClick(input services.RecordClickInput) {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), recordClickTimeout)
		defer cancel()

		if err := h.service.Clicks.RecordClick(ctx, input); err != nil {
			log.Printf("failed to record click for short_url_id=%d: %v\n", input.ShortURLID, err)
		}
	}()
}

// clientIP prefers the first X-Forwarded-For entry when present, falling back
// to the connection's remote address. XFF is client-controlled, so it should
// only be trusted when the service sits behind a proxy that sets it.
func clientIP(r *http.Request) string {
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		first := strings.TrimSpace(strings.Split(forwarded, ",")[0])
		if first != "" {
			return first
		}
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}

	return host
}
