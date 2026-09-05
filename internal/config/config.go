package config

import (
	"errors"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	DbConfig     *DbConfig
}

func NewConfig(port string, readTimeout, writeTimeout time.Duration, dbConfig *DbConfig) *Config {
	return &Config{
		Port:         port,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		DbConfig:     dbConfig,
	}
}

type DbConfig struct {
	Url             string
	MaxOpenConns    int
	MaxIdleConns    int
	MaxConnLifetime time.Duration
	MaxConnIdleTime time.Duration
}

func NewDbConfig(url string, maxOpenConns int, maxIdleConns int, maxConnLifetime time.Duration, maxConnIdleTime time.Duration) *DbConfig {
	return &DbConfig{
		Url:             url,
		MaxOpenConns:    maxOpenConns,
		MaxIdleConns:    maxIdleConns,
		MaxConnLifetime: maxConnLifetime,
		MaxConnIdleTime: maxConnIdleTime,
	}
}

func LoadConfig() (*Config, error) {

	_ = godotenv.Load()

	port := os.Getenv("PORT")
	databaseUrl := os.Getenv("DATABASE_URL")

	if port == "" || databaseUrl == "" {
		return nil, errors.New("$PORT or $DATABASE_URL is not set")
	}

	readTimeout := 15 * time.Second
	writeTimeout := 15 * time.Second
	maxOpenConns := 25
	maxIdleConns := 10
	maxConnLifetime := 5 * time.Minute
	maxConnIdleTime := 30 * time.Minute

	dbConfig := NewDbConfig(databaseUrl, maxOpenConns, maxIdleConns, maxConnLifetime, maxConnIdleTime)
	return NewConfig(port, readTimeout, writeTimeout, dbConfig), nil
}
