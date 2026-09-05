package handlers

import (
	"net/http"

	"github.com/dhruv15803/url-shortener-service/internal/services"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	Users IUserHandler
	Auth  IAuthHandler
}

func NewHandler(service *services.Service) *Handler {
	return &Handler{
		Users: NewUserHandler(service),
		Auth:  NewAuthHandler(service),
	}
}

// RegisterRoutes mounts every sub-handler's routes onto r.
func (h *Handler) RegisterRoutes(r chi.Router) {
	h.Users.RegisterRoutes(r)
	h.Auth.RegisterRoutes(r)
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
