.PHONY: help setup db-up db-down migrate-up migrate-down sqlc build run test clean

help:
	@echo "Available commands:"
	@echo "  setup        - Install dependencies and tools"
	@echo "  db-up        - Start PostgreSQL container"
	@echo "  db-down      - Stop PostgreSQL container"
	@echo "  migrate-up   - Run database migrations"
	@echo "  migrate-down - Rollback database migrations"
	@echo "  sqlc         - Generate sqlc code"
	@echo "  build        - Build the application"
	@echo "  run          - Run the application"
	@echo "  test         - Run tests"
	@echo "  clean        - Clean build artifacts"

setup:
	go mod download
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

db-up:
	docker-compose up -d postgres

db-down:
	docker-compose down

migrate-up:
	migrate -path migrations -database "postgres://allocation_user:allocation_pass@localhost:54321/allocation_db?sslmode=disable" up

migrate-down:
	migrate -path migrations -database "postgres://allocation_user:allocation_pass@localhost:54321/allocation_db?sslmode=disable" down

sqlc:
	sqlc generate

build:
	go build -o bin/server ./cmd/server

run:
	go run ./cmd/server

test:
	go test -v ./...

clean:
	rm -rf bin/
