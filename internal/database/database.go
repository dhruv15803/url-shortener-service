package database

import (
	"time"

	"github.com/jmoiron/sqlx"
)

type Database interface {
	Connect() (*sqlx.DB, error)
}

func NewDatabase(url string, maxOpenConns int, maxIdleConns int, maxConnLifetime time.Duration, maxConnIdleTime time.Duration) Database {
	return NewPostgresDatabase(url, maxOpenConns, maxIdleConns, maxConnLifetime, maxConnIdleTime)
}
