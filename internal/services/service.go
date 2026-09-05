package services

import "github.com/dhruv15803/url-shortener-service/internal/repositories"

type Service struct {
	Users IUserService
}

func NewService(repository *repositories.Repository) *Service {
	return &Service{
		Users: NewUserService(repository),
	}
}

type IUserService interface {
	DeleteUserByID(id int) error
}
