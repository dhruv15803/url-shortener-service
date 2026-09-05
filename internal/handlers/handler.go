package handlers

import (
	"net/http"

	"github.com/dhruv15803/url-shortener-service/internal/services"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	Users IUserHandler
}

func NewHandler(service *services.Service) *Handler {
	return &Handler{
		Users: NewUserHandler(service),
	}
}

// RegisterRoutes mounts every sub-handler's routes onto r.
func (h *Handler) RegisterRoutes(r chi.Router) {
	h.Users.RegisterRoutes(r)
}

type IUserHandler interface {
	DeleteUserByID(w http.ResponseWriter, r *http.Request)
	RegisterRoutes(r chi.Router)
}
