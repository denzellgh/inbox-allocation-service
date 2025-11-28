# ============================================
# Inbox Allocation Service - Makefile
# ============================================

# Build variables
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ 2>/dev/null || echo "unknown")
GIT_COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DOCKER_REGISTRY ?= ghcr.io/your-org

# Go variables
GOCMD = go
GOBUILD = $(GOCMD) build
GOTEST = $(GOCMD) test
GOGET = $(GOCMD) get
GOMOD = $(GOCMD) mod
BINARY_NAME = server
BINARY_PATH = bin/$(BINARY_NAME)

# Build flags
LDFLAGS = -ldflags "-w -s \
	-X main.Version=$(VERSION) \
	-X main.BuildTime=$(BUILD_TIME) \
	-X main.GitCommit=$(GIT_COMMIT)"

.PHONY: help
help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# ============================================
# Development
# ============================================

.PHONY: setup
setup: ## Install development dependencies
	$(GOGET) github.com/sqlc-dev/sqlc/cmd/sqlc@latest
	$(GOGET) -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	$(GOGET) github.com/air-verse/air@latest
	$(GOMOD) download

.PHONY: run
run: ## Run the application locally
	$(GOCMD) run ./cmd/server

.PHONY: run-dev
run-dev: ## Run with hot reload (requires air)
	air -c .air.toml

.PHONY: build
build: ## Build the application
	CGO_ENABLED=0 $(GOBUILD) $(LDFLAGS) -o $(BINARY_PATH) ./cmd/server

.PHONY: clean
clean: ## Clean build artifacts
	rm -rf bin/
	rm -f coverage coverage.out coverage.html

# ============================================
# Database
# ============================================

.PHONY: db-up
db-up: ## Start PostgreSQL container
	docker-compose up -d postgres
	@echo "Waiting for PostgreSQL to be ready..."
	@timeout /t 5 /nobreak >nul 2>&1 || sleep 5

.PHONY: db-down
db-down: ## Stop PostgreSQL container
	docker-compose down

.PHONY: db-logs
db-logs: ## View PostgreSQL logs
	docker-compose logs -f postgres

.PHONY: db-shell
db-shell: ## Connect to PostgreSQL shell
	docker exec -it allocation_postgres psql -U allocation_user -d allocation_db

.PHONY: migrate-up
migrate-up: ## Run database migrations
	migrate -path migrations -database "postgres://allocation_user:allocation_pass@localhost:54321/allocation_db?sslmode=disable" up

.PHONY: migrate-down
migrate-down: ## Rollback last migration
	migrate -path migrations -database "postgres://allocation_user:allocation_pass@localhost:54321/allocation_db?sslmode=disable" down 1

.PHONY: migrate-reset
migrate-reset: ## Reset all migrations
	migrate -path migrations -database "postgres://allocation_user:allocation_pass@localhost:54321/allocation_db?sslmode=disable" drop -f

.PHONY: sqlc
sqlc: ## Generate sqlc code
	sqlc generate

# ============================================
# Testing
# ============================================

.PHONY: test
test: ## Run unit tests
	$(GOTEST) -v -short ./...

.PHONY: test-race
test-race: ## Run tests with race detector
	$(GOTEST) -v -race ./...

.PHONY: test-integration
test-integration: ## Run integration tests (requires Docker)
	$(GOTEST) -v -tags=integration ./...

.PHONY: test-all
test-all: ## Run all tests
	$(GOTEST) -v -tags=integration -race ./...

.PHONY: coverage
coverage: ## Generate coverage report
	$(GOTEST) -v -short -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

.PHONY: bench
bench: ## Run benchmarks
	$(GOTEST) -bench=. -benchmem ./...

# ============================================
# Docker
# ============================================

.PHONY: docker-build
docker-build: ## Build Docker image
	docker build \
		--build-arg VERSION=$(VERSION) \
		--build-arg BUILD_TIME=$(BUILD_TIME) \
		--build-arg GIT_COMMIT=$(GIT_COMMIT) \
		-t inbox-allocation-service:$(VERSION) \
		-t inbox-allocation-service:latest \
		.

.PHONY: docker-build-dev
docker-build-dev: ## Build development Docker image
	docker build \
		--target development \
		-t inbox-allocation-service:dev \
		.

.PHONY: docker-run-dev
docker-run-dev: ## Run development stack with hot reload
	docker-compose up -d

.PHONY: docker-stop
docker-stop: ## Stop Docker Compose
	docker-compose down

.PHONY: docker-logs
docker-logs: ## View application logs
	docker-compose logs -f app

.PHONY: docker-shell
docker-shell: ## Shell into running container
	docker exec -it allocation_app /bin/sh

.PHONY: docker-clean
docker-clean: ## Clean Docker resources
	docker-compose down -v --remove-orphans
	docker image prune -f

.PHONY: docker-push
docker-push: docker-build ## Push to registry
	docker tag inbox-allocation-service:$(VERSION) $(DOCKER_REGISTRY)/inbox-allocation-service:$(VERSION)
	docker tag inbox-allocation-service:latest $(DOCKER_REGISTRY)/inbox-allocation-service:latest
	docker push $(DOCKER_REGISTRY)/inbox-allocation-service:$(VERSION)
	docker push $(DOCKER_REGISTRY)/inbox-allocation-service:latest

# ============================================
# Production
# ============================================

.PHONY: prod-up
prod-up: ## Start production stack
	docker-compose -f docker-compose.prod.yml up -d

.PHONY: prod-down
prod-down: ## Stop production stack
	docker-compose -f docker-compose.prod.yml down

.PHONY: prod-logs
prod-logs: ## View production logs
	docker-compose -f docker-compose.prod.yml logs -f

.PHONY: prod-migrate
prod-migrate: ## Run production migrations
	docker-compose -f docker-compose.prod.yml --profile migrate run --rm migrate

.PHONY: prod-scale
prod-scale: ## Scale production replicas (usage: make prod-scale REPLICAS=3)
	docker-compose -f docker-compose.prod.yml up -d --scale app=$(REPLICAS)

# ============================================
# Quick Start
# ============================================

.PHONY: dev-up
dev-up: db-up ## Start full development environment
	@echo "Waiting for database..."
	@timeout /t 5 /nobreak >nul 2>&1 || sleep 5
	$(MAKE) migrate-up
	$(MAKE) run

.PHONY: dev-down
dev-down: ## Stop development environment
	docker-compose down -v

.PHONY: fresh
fresh: dev-down db-up ## Fresh start (reset everything)
	@timeout /t 5 /nobreak >nul 2>&1 || sleep 5
	$(MAKE) migrate-up
	$(MAKE) sqlc
	$(MAKE) run

# ============================================
# Utilities
# ============================================

.PHONY: fmt
fmt: ## Format code
	$(GOCMD) fmt ./...

.PHONY: lint
lint: ## Run linter
	golangci-lint run

.PHONY: tidy
tidy: ## Tidy go modules
	$(GOMOD) tidy

.PHONY: verify
verify: fmt lint test ## Verify code (format, lint, test)

.PHONY: version
version: ## Show version info
	@echo "Version:    $(VERSION)"
	@echo "Build Time: $(BUILD_TIME)"
	@echo "Git Commit: $(GIT_COMMIT)"
