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
	GeoIPDBPath     string
	DbConfig        *DbConfig
	OAuthConfig     *OAuthConfig
	JWTConfig       *JWTConfig
	RedisConfig     *RedisConfig
}

func NewConfig(port string, readTimeout, writeTimeout time.Duration, shortURLBaseURL string, geoIPDBPath string, dbConfig *DbConfig, oauthConfig *OAuthConfig, jwtConfig *JWTConfig, redisConfig *RedisConfig) *Config {
	return &Config{
		Port:            port,
		ReadTimeout:     readTimeout,
		WriteTimeout:    writeTimeout,
		ShortURLBaseURL: shortURLBaseURL,
		GeoIPDBPath:     geoIPDBPath,
		DbConfig:        dbConfig,
		OAuthConfig:     oauthConfig,
		JWTConfig:       jwtConfig,
		RedisConfig:     redisConfig,
	}
}

// WorkerConfig is the narrower config the click worker needs. It deliberately
// omits the http/oauth/jwt settings so the worker doesn't require credentials
// for things it never does.
type WorkerConfig struct {
	DbConfig    *DbConfig
	RedisConfig *RedisConfig
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

func NewRedisConfig(addr string, password string, db int) *RedisConfig {
	return &RedisConfig{
		Addr:     addr,
		Password: password,
		DB:       db,
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
	geoIPDBPath := os.Getenv("GEOIP_DB_PATH")
	redisAddr := os.Getenv("REDIS_ADDR")

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

	if redisAddr == "" {
		return nil, errors.New("$REDIS_ADDR is not set")
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
	redisConfig := NewRedisConfig(redisAddr, os.Getenv("REDIS_PASSWORD"), 0)
	return NewConfig(port, readTimeout, writeTimeout, shortURLBaseURL, geoIPDBPath, dbConfig, oauthConfig, jwtConfig, redisConfig), nil
}

// LoadWorkerConfig loads only what the click worker needs - the database it
// writes to and the redis queue it consumes from.
func LoadWorkerConfig() (*WorkerConfig, error) {

	_ = godotenv.Load()

	databaseUrl := os.Getenv("DATABASE_URL")
	redisAddr := os.Getenv("REDIS_ADDR")

	if databaseUrl == "" {
		return nil, errors.New("$DATABASE_URL is not set")
	}

	if redisAddr == "" {
		return nil, errors.New("$REDIS_ADDR is not set")
	}

	maxOpenConns := 25
	maxIdleConns := 10
	maxConnLifetime := 5 * time.Minute
	maxConnIdleTime := 30 * time.Minute

	return &WorkerConfig{
		DbConfig:    NewDbConfig(databaseUrl, maxOpenConns, maxIdleConns, maxConnLifetime, maxConnIdleTime),
		RedisConfig: NewRedisConfig(redisAddr, os.Getenv("REDIS_PASSWORD"), 0),
	}, nil
}
