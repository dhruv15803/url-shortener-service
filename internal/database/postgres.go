package database

import (
	"time"

	"github.com/jmoiron/sqlx"
)

type PostgresDatabase struct {
	Url             string
	MaxOpenConns    int
	MaxIdleConns    int
	MaxConnLifetime time.Duration
	MaxConnIdleTime time.Duration
}

func NewPostgresDatabase(url string, maxOpenConns int, maxIdleConns int, maxConnLifetime time.Duration, maxConnIdleTime time.Duration) *PostgresDatabase {
	return &PostgresDatabase{
		Url:             url,
		MaxOpenConns:    maxOpenConns,
		MaxIdleConns:    maxIdleConns,
		MaxConnLifetime: maxConnLifetime,
		MaxConnIdleTime: maxConnIdleTime,
	}
}

func (p *PostgresDatabase) Connect() (*sqlx.DB, error) {

	db, err := sqlx.Open("postgres", p.Url)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(p.MaxOpenConns)
	db.SetMaxIdleConns(p.MaxIdleConns)
	db.SetConnMaxLifetime(p.MaxConnLifetime)
	db.SetConnMaxIdleTime(p.MaxConnIdleTime)

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}
