package main

import (
	"log"
	"net/http"
	"time"

	"github.com/dhruv15803/url-shortener-service/internal/cache"
	"github.com/dhruv15803/url-shortener-service/internal/config"
	"github.com/dhruv15803/url-shortener-service/internal/database"
	"github.com/dhruv15803/url-shortener-service/internal/geoip"
	"github.com/dhruv15803/url-shortener-service/internal/handlers"
	"github.com/dhruv15803/url-shortener-service/internal/queue"
	"github.com/dhruv15803/url-shortener-service/internal/repositories"
	"github.com/dhruv15803/url-shortener-service/internal/services"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/redis/go-redis/v9"
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

	log.Printf("connected to db!\n")

	geoipReader, err := geoip.NewReader(cfg.GeoIPDBPath)
	if err != nil {
		log.Fatalf("failed to open geoip database at %v: %v\n", cfg.GeoIPDBPath, err)
	}
	defer geoipReader.Close()

	if geoipReader.Enabled() {
		log.Printf("geoip enabled: %v\n", cfg.GeoIPDBPath)
	} else {
		log.Printf("geoip disabled ($GEOIP_DB_PATH not set), clicks will have no location data\n")
	}

	// Short timeouts and no retries: redis is on the redirect hot path, so an
	// unreachable redis must fail fast and let the request fall through to
	// postgres. The client defaults (5s dial, 3 retries) would otherwise turn
	// a redis outage into multi-second redirects.
	redisClient := redis.NewClient(&redis.Options{
		Addr:         cfg.RedisConfig.Addr,
		Password:     cfg.RedisConfig.Password,
		DB:           cfg.RedisConfig.DB,
		DialTimeout:  200 * time.Millisecond,
		ReadTimeout:  200 * time.Millisecond,
		WriteTimeout: 200 * time.Millisecond,
		MaxRetries:   -1,
	})
	defer redisClient.Close()

	clickQueue := queue.NewClickQueue(redisClient)
	shortURLCache := cache.NewShortURLCache(redisClient)
	analyticsCache := cache.NewAnalyticsCache(redisClient)

	repository := repositories.NewRepository(db)
	service := services.NewService(repository, cfg, clickQueue, geoipReader, shortURLCache, analyticsCache)
	handler := handlers.NewHandler(service, cfg)

	r := chi.NewRouter()

	// Must be registered before any routes: chi only applies middleware added
	// ahead of route registration on a given mux. The origin has to be exact
	// rather than "*", because the session cookie makes these credentialed
	// requests.
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{cfg.FrontendURL},
		AllowedMethods:   []string{"GET", "POST", "PUT", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Route("/api", func(r chi.Router) {
		r.Get("/health", func(w http.ResponseWriter, r *http.Request) {

			w.Write([]byte("hello world"))

		})

		handler.RegisterRoutes(r)
	})

	// Registered after /api so GET /{shortCode} can't shadow the api routes.
	handler.RegisterRootRoutes(r)

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
