package services

import "github.com/dhruv15803/url-shortener-service/internal/repositories"

// UserService holds the full repository aggregate (not just repository.Users)
// so its business logic can reach into other domains' repositories when needed.
type UserService struct {
	repository *repositories.Repository
}

func NewUserService(repository *repositories.Repository) *UserService {
	return &UserService{
		repository: repository,
	}
}

func (s *UserService) DeleteUserByID(id int) error {
	return s.repository.Users.DeleteUserByID(id)
}
