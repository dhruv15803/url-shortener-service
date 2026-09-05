package config

import (
	"errors"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Port            string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShortURLBaseURL string
	DbConfig        *DbConfig
	OAuthConfig     *OAuthConfig
	JWTConfig       *JWTConfig
}

func NewConfig(port string, readTimeout, writeTimeout time.Duration, shortURLBaseURL string, dbConfig *DbConfig, oauthConfig *OAuthConfig, jwtConfig *JWTConfig) *Config {
	return &Config{
		Port:            port,
		ReadTimeout:     readTimeout,
		WriteTimeout:    writeTimeout,
		ShortURLBaseURL: shortURLBaseURL,
		DbConfig:        dbConfig,
		OAuthConfig:     oauthConfig,
		JWTConfig:       jwtConfig,
	}
}

type OAuthConfig struct {
	GoogleClientID     string
	GoogleClientSecret string
	GoogleRedirectURL  string
}

func NewOAuthConfig(googleClientID, googleClientSecret, googleRedirectURL string) *OAuthConfig {
	return &OAuthConfig{
		GoogleClientID:     googleClientID,
		GoogleClientSecret: googleClientSecret,
		GoogleRedirectURL:  googleRedirectURL,
	}
}

type JWTConfig struct {
	Secret string
	Expiry time.Duration
}

func NewJWTConfig(secret string, expiry time.Duration) *JWTConfig {
	return &JWTConfig{
		Secret: secret,
		Expiry: expiry,
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
	googleClientID := os.Getenv("GOOGLE_CLIENT_ID")
	googleClientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")
	googleRedirectURL := os.Getenv("GOOGLE_REDIRECT_URL")
	jwtSecret := os.Getenv("JWT_SECRET")
	shortURLBaseURL := os.Getenv("SHORT_URL_BASE_URL")

	if port == "" || databaseUrl == "" {
		return nil, errors.New("$PORT or $DATABASE_URL is not set")
	}

	if googleClientID == "" || googleClientSecret == "" || googleRedirectURL == "" {
		return nil, errors.New("$GOOGLE_CLIENT_ID or $GOOGLE_CLIENT_SECRET or $GOOGLE_REDIRECT_URL is not set")
	}

	if jwtSecret == "" {
		return nil, errors.New("$JWT_SECRET is not set")
	}

	if shortURLBaseURL == "" {
		return nil, errors.New("$SHORT_URL_BASE_URL is not set")
	}

	readTimeout := 15 * time.Second
	writeTimeout := 15 * time.Second
	maxOpenConns := 25
	maxIdleConns := 10
	maxConnLifetime := 5 * time.Minute
	maxConnIdleTime := 30 * time.Minute
	jwtExpiry := 24 * time.Hour

	dbConfig := NewDbConfig(databaseUrl, maxOpenConns, maxIdleConns, maxConnLifetime, maxConnIdleTime)
	oauthConfig := NewOAuthConfig(googleClientID, googleClientSecret, googleRedirectURL)
	jwtConfig := NewJWTConfig(jwtSecret, jwtExpiry)
	return NewConfig(port, readTimeout, writeTimeout, shortURLBaseURL, dbConfig, oauthConfig, jwtConfig), nil
}
