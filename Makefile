DATABASE_URL := postgresql://postgres:1234%40%23A@localhost:5432/url-shortener-db?sslmode=disable
MIGRATIONS_DIR := ./internal/database/migrations

.PHONY: run worker migrate-create migrate-up migrate-down

# Start API server
run:
	go run ./cmd/api

# Start the background click worker (consumes the redis queue)
worker:
	go run ./cmd/worker

# Create a new migration
# Usage: make migrate-create name=create_users_table
migrate-create:
	migrate create -ext sql -dir $(MIGRATIONS_DIR) -seq $(name)

# Run all pending migrations
migrate-up:
	migrate -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" up

# Roll back the last migration
migrate-down:
	migrate -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" down 1