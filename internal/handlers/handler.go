package handlers

import (
	"net/http"

	"github.com/dhruv15803/url-shortener-service/internal/config"
	"github.com/dhruv15803/url-shortener-service/internal/services"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	Users    IUserHandler
	Auth     IAuthHandler
	Urls     IUrlHandler
	Redirect IRedirectHandler
}

func NewHandler(service *services.Service, cfg *config.Config) *Handler {
	return &Handler{
		Users:    NewUserHandler(service),
		Auth:     NewAuthHandler(service),
		Urls:     NewUrlHandler(service, cfg.JWTConfig.Secret, cfg.ShortURLBaseURL),
		Redirect: NewRedirectHandler(service),
	}
}

// RegisterRoutes mounts every sub-handler's routes onto r (the /api router).
func (h *Handler) RegisterRoutes(r chi.Router) {
	h.Users.RegisterRoutes(r)
	h.Auth.RegisterRoutes(r)
	h.Urls.RegisterRoutes(r)
}

// RegisterRootRoutes mounts routes that live at the root rather than under
// /api - short links have to be short.
func (h *Handler) RegisterRootRoutes(r chi.Router) {
	h.Redirect.RegisterRoutes(r)
}

type IUserHandler interface {
	DeleteUserByID(w http.ResponseWriter, r *http.Request)
	RegisterRoutes(r chi.Router)
}

type IAuthHandler interface {
	GoogleLogin(w http.ResponseWriter, r *http.Request)
	GoogleCallback(w http.ResponseWriter, r *http.Request)
	RegisterRoutes(r chi.Router)
}

type IUrlHandler interface {
	CreateOrGetDestination(w http.ResponseWriter, r *http.Request)
	CreateShortURLForDestination(w http.ResponseWriter, r *http.Request)
	ListCampaigns(w http.ResponseWriter, r *http.Request)
	GetCampaign(w http.ResponseWriter, r *http.Request)
	UpdateCampaign(w http.ResponseWriter, r *http.Request)
	RegisterRoutes(r chi.Router)
}

type IRedirectHandler interface {
	Redirect(w http.ResponseWriter, r *http.Request)
	RegisterRoutes(r chi.Router)
}
