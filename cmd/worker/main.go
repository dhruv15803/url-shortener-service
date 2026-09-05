package main

import (
	"context"
	"log"

	"github.com/dhruv15803/url-shortener-service/internal/config"
	"github.com/dhruv15803/url-shortener-service/internal/database"
	"github.com/dhruv15803/url-shortener-service/internal/geoip"
	"github.com/dhruv15803/url-shortener-service/internal/queue"
	"github.com/dhruv15803/url-shortener-service/internal/repositories"
	"github.com/dhruv15803/url-shortener-service/internal/services"
	"github.com/redis/go-redis/v9"
)

func main() {

	cfg, err := config.LoadWorkerConfig()
	if err != nil {
		log.Fatalf("error loading worker config: %v\n", err)
	}

	db, err := database.NewDatabase(cfg.DbConfig.Url, cfg.DbConfig.MaxOpenConns, cfg.DbConfig.MaxIdleConns, cfg.DbConfig.MaxConnLifetime, cfg.DbConfig.MaxConnIdleTime).Connect()
	if err != nil {
		log.Fatalf("connection to database failed: %v\n", err)
	}
	defer db.Close()

	log.Printf("connected to db!\n")

	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisConfig.Addr,
		Password: cfg.RedisConfig.Password,
		DB:       cfg.RedisConfig.DB,
	})
	defer redisClient.Close()

	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("connection to redis failed: %v\n", err)
	}

	log.Printf("connected to redis!\n")

	// The worker never enriches - events arrive already parsed from the api -
	// so it runs with a disabled geoip reader and doesn't need the mmdb.
	geoipReader, err := geoip.NewReader("")
	if err != nil {
		log.Fatalf("failed to build geoip reader: %v\n", err)
	}

	repository := repositories.NewRepository(db)
	clickQueue := queue.NewClickQueue(redisClient)
	clickService := services.NewClickService(repository, clickQueue, geoipReader)

	log.Printf("click worker started, consuming %v\n", queue.ClickQueueKey)

	clickQueue.Consume(context.Background(), clickService.PersistClick)
}
