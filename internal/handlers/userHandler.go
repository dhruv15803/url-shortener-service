package handlers

import (
	"net/http"
	"strconv"

	"github.com/dhruv15803/url-shortener-service/internal/httpresponse"
	"github.com/dhruv15803/url-shortener-service/internal/services"
	"github.com/go-chi/chi/v5"
)

// UserHandler holds the full service aggregate (not just service.Users) so it
// can call other domains' services if an endpoint's logic ever needs to.
type UserHandler struct {
	service *services.Service
}

func NewUserHandler(service *services.Service) *UserHandler {
	return &UserHandler{
		service: service,
	}
}

func (h *UserHandler) RegisterRoutes(r chi.Router) {
	r.Delete("/users/{id}", h.DeleteUserByID)
}

func (h *UserHandler) DeleteUserByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		httpresponse.WriteError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	if err := h.service.Users.DeleteUserByID(id); err != nil {
		httpresponse.WriteError(w, http.StatusInternalServerError, "failed to delete user")
		return
	}

	httpresponse.WriteJSON(w, http.StatusOK, map[string]string{"message": "user deleted"})
}
