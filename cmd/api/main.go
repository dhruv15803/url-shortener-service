package main

import (
	"log"
	"net/http"

	"github.com/dhruv15803/url-shortener-service/internal/config"
	"github.com/dhruv15803/url-shortener-service/internal/database"
	"github.com/dhruv15803/url-shortener-service/internal/handlers"
	"github.com/dhruv15803/url-shortener-service/internal/repositories"
	"github.com/dhruv15803/url-shortener-service/internal/services"
	"github.com/go-chi/chi/v5"
)

func main() {

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("error loading config: %v\n", err)
	}

	db, err := database.NewDatabase(cfg.DbConfig.Url, cfg.DbConfig.MaxOpenConns, cfg.DbConfig.MaxIdleConns, cfg.DbConfig.MaxConnLifetime, cfg.DbConfig.MaxConnIdleTime).Connect()
	if err != nil {
		log.Fatalf("connection to database failed: %v\n", err)
	}

	repository := repositories.NewRepository(db)
	service := services.NewService(repository)
	handler := handlers.NewHandler(service)

	r := chi.NewRouter()

	r.Route("/api", func(r chi.Router) {
		r.Get("/health", func(w http.ResponseWriter, r *http.Request) {

			w.Write([]byte("hello world"))

		})

		handler.RegisterRoutes(r)
	})

	server := http.Server{
		Addr:         ":" + cfg.Port,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		Handler:      r,
	}

	log.Printf("Starting server on port: %v\n", cfg.Port)

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("failed to start server on port: %v\n", err)
	}
}
