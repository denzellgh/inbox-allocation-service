# Inbox Allocation Service - Backend

High-performance microservice for inbox management and conversation allocation built with Go, Chi, and PostgreSQL.

> [!IMPORTANT]
>
> This project is in active development. Please use the `main` branch for stable code. Other branches like `dev` and `staging` contain new features and changes that are currently being tested. If you have any suggestions or feature requests, feel free to open an issue on GitHub.

## Quick Start

```bash
# 1. Clone and install dependencies
cd backend
go mod download

# 2. Start PostgreSQL
make db-up

# 3. Run migrations
make migrate-up

# 4. Start the service
make run
```

The service will be available at `http://localhost:8080`

## Table of Contents

- [Features](#features)
- [Architecture](#architecture)
- [Requirements](#requirements)
- [Installation](#installation)
- [Configuration](#configuration)
- [API Documentation](#api-documentation)
- [Development](#development)
- [Testing](#testing)
- [Deployment](#deployment)
- [Troubleshooting](#troubleshooting)

## Features

### Core Capabilities
- **Auto-allocation**: Priority-based automatic conversation assignment
- **Manual Claim**: Operators can claim specific conversations
- **Grace Period**: Configurable grace period when operators go offline
- **Labels**: Per-inbox labels for conversation organization
- **Multi-tenancy**: Strict tenant isolation at database level
- **Idempotency**: Safe retry operations with idempotency keys

### Technical Features
- **Concurrency Safety**: Row-level locking with `FOR UPDATE SKIP LOCKED`
- **Structured Logging**: Context-aware logging with correlation IDs
- **Graceful Shutdown**: Clean shutdown with resource cleanup hooks
- **Connection Pooling**: Health-monitored database connection pool
- **Retry Logic**: Exponential backoff with jitter for transient failures
- **Observability**: Request tracing and performance monitoring

## Architecture

The service follows Clean Architecture principles with clear separation of concerns:

```
HTTP Layer → API Handlers → Service Layer → Repository Layer → PostgreSQL
```

For detailed architecture documentation, see [ARCHITECTURE.md](./ARCHITECTURE.md)

### Database Schema

**9 Core Tables:**
1. `tenants` - Tenant configuration with priority weights
2. `inboxes` - Communication channels (phone numbers)
3. `operators` - System users
4. `operator_inbox_subscriptions` - Operator-inbox subscriptions
5. `operator_status` - Real-time operator availability
6. `conversation_refs` - Conversation metadata
7. `labels` - Per-inbox labels
8. `conversation_labels` - Conversation-label relationships
9. `grace_period_assignments` - Grace period tracking

**Critical Indexes:**
- `idx_conversations_allocation` - Optimized for `FOR UPDATE SKIP LOCKED`
- `idx_grace_expires` - For grace period worker
- All composite indexes start with `tenant_id` (multi-tenancy)

## Requirements

- **Go**: 1.22 or higher
- **Docker**: 20.10+ and Docker Compose
- **Make**: For development commands (optional)
- **Port**: 54321 available for PostgreSQL

## Installation

### 1. Install Dependencies

```bash
cd backend
go mod download
```

### 2. Install Development Tools

```bash
make setup
```

Or manually:

```bash
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

### 3. Configure Environment

```bash
cp .env.example .env
```

Edit `.env` according to your needs.

## Configuration

### Environment Variables

```bash
# Server Configuration
SERVER_HOST=0.0.0.0
SERVER_PORT=8080
SERVER_READ_TIMEOUT=15s
SERVER_WRITE_TIMEOUT=15s
SERVER_SHUTDOWN_TIMEOUT=30s

# Database Configuration
DB_HOST=127.0.0.1
DB_PORT=54321
DB_USER=allocation_user
DB_PASSWORD=allocation_pass
DB_NAME=allocation_db
DB_MAX_CONNS=25
DB_MIN_CONNS=5

# Logging
LOG_LEVEL=info        # debug, info, warn, error
LOG_FORMAT=json       # json, console

# Workers
WORKER_GRACE_PERIOD_INTERVAL=30s
WORKER_GRACE_PERIOD_BATCH_SIZE=100

# Idempotency
IDEMPOTENCY_TTL=24h
IDEMPOTENCY_CLEANUP_INTERVAL=1h
```

## API Documentation

### Interactive Documentation

When the service is running, access the interactive Swagger UI at:

**🔗 http://localhost:8080/docs**

This provides:
- Interactive API testing
- Request/response examples
- Schema definitions
- Authentication setup

### OpenAPI Specification

The complete API specification is also available at:
- **File**: [`api/openapi.yaml`](./api/openapi.yaml)
- **Endpoint**: `http://localhost:8080/api/openapi.yaml`

### Authentication

All API routes under `/api/v1` require headers:
- `X-Tenant-ID`: Tenant UUID (required)
- `X-Operator-ID`: Operator UUID (required for protected routes)

### Idempotency

Mutation endpoints support the `Idempotency-Key` header for safe retries:

```bash
curl -X POST http://localhost:8080/api/v1/allocate \
  -H "X-Tenant-ID: 550e8400-e29b-41d4-a716-446655440000" \
  -H "X-Operator-ID: 660e8400-e29b-41d4-a716-446655440001" \
  -H "Idempotency-Key: unique-key-123"
```

### Example Requests

**Get Operator Status:**
```bash
curl http://localhost:8080/api/v1/operator/status \
  -H "X-Tenant-ID: <tenant-uuid>" \
  -H "X-Operator-ID: <operator-uuid>"
```

**Auto-allocate Conversation (no body required):**
```bash
curl -X POST http://localhost:8080/api/v1/allocate \
  -H "X-Tenant-ID: <tenant-uuid>" \
  -H "X-Operator-ID: <operator-uuid>"
```

**Manually Claim Conversation:**
```bash
curl -X POST http://localhost:8080/api/v1/claim \
  -H "X-Tenant-ID: <tenant-uuid>" \
  -H "X-Operator-ID: <operator-uuid>" \
  -H "Content-Type: application/json" \
  -d '{"conversation_id": "<conversation-uuid>"}'
```

**List Conversations:**
```bash
curl "http://localhost:8080/api/v1/conversations?state=QUEUED&limit=50" \
  -H "X-Tenant-ID: <tenant-uuid>" \
  -H "X-Operator-ID: <operator-uuid>"
```

**Resolve Conversation:**
```bash
curl -X POST http://localhost:8080/api/v1/resolve \
  -H "X-Tenant-ID: <tenant-uuid>" \
  -H "X-Operator-ID: <operator-uuid>" \
  -H "Content-Type: application/json" \
  -d '{"conversation_id": "<conversation-uuid>"}'
```

**Subscribe Operator to Inbox:**
```bash
curl -X POST http://localhost:8080/api/v1/inboxes/<inbox-uuid>/operators \
  -H "X-Tenant-ID: <tenant-uuid>" \
  -H "Content-Type: application/json" \
  -d '{"operator_id": "<operator-uuid>"}'
```

**Attach Label to Conversation:**
```bash
curl -X POST http://localhost:8080/api/v1/labels/attach \
  -H "X-Tenant-ID: <tenant-uuid>" \
  -H "X-Operator-ID: <operator-uuid>" \
  -H "Content-Type: application/json" \
  -d '{"label_id": "<label-uuid>", "conversation_id": "<conversation-uuid>"}'
```

## Development

### Available Commands

```bash
make help          # Show help
make setup         # Install dependencies and tools
make db-up         # Start PostgreSQL
make db-down       # Stop PostgreSQL
make migrate-up    # Run migrations
make migrate-down  # Rollback migrations
make sqlc          # Generate sqlc code
make build         # Build application
make run           # Run application
make test          # Run tests
make lint          # Run linters
make clean         # Clean artifacts
```

### Database Management

**Start Database:**
```bash
make db-up
# or
docker-compose up -d postgres
```

**Run Migrations:**
```bash
make migrate-up
```

**Verify Tables:**
```bash
docker exec allocation_postgres psql -U allocation_user -d allocation_db -c "\dt"
```

**Generate sqlc Code:**
```bash
make sqlc
```

### Running the Service

**Development Mode:**
```bash
make run
# or
go run ./cmd/server
```

**Build Binary:**
```bash
make build
./bin/server
```

## Project Structure

```
backend/
├── cmd/
│   └── server/              # Application entry point
├── internal/
│   ├── api/                 # HTTP layer
│   │   ├── handlers/        # Request handlers
│   │   ├── middleware/      # HTTP middleware
│   │   ├── dto/             # Data transfer objects
│   │   └── router.go        # Route definitions
│   ├── service/             # Business logic layer
│   ├── repository/          # Data access layer (sqlc generated)
│   ├── domain/              # Domain models
│   ├── config/              # Configuration
│   ├── pkg/                 # Shared packages
│   │   ├── logger/          # Structured logging
│   │   ├── database/        # DB utilities
│   │   └── retry/           # Retry logic
│   ├── server/              # HTTP server
│   └── worker/              # Background workers
├── migrations/              # Database migrations
├── api/                     # API documentation
│   └── openapi.yaml         # OpenAPI 3.0 spec
├── docker-compose.yml       # Docker services
├── Makefile                 # Development commands
└── sqlc.yaml                # sqlc configuration
```

## Testing

### Run Tests

```bash
# All tests
make test

# With coverage
go test -v -cover ./...

# Specific package
go test -v ./internal/service/...
```

### Test Database

For integration tests, use a separate test database:

```bash
docker-compose up -d postgres-test
make migrate-up-test
```

## Deployment

### Health Checks

The service exposes health check endpoints:

- `GET /health` - Liveness probe (always returns 200)
- `GET /ready` - Readiness probe (checks DB connection)
- `GET /version` - Version and build information

### Graceful Shutdown

The service handles `SIGINT` and `SIGTERM` signals:

1. Stops accepting new requests
2. Waits for in-flight requests (max 30s)
3. Stops background workers
4. Closes database connections
5. Exits cleanly

### Docker Build

```bash
docker build -t inbox-allocation-service:latest .
docker run -p 8080:8080 --env-file .env inbox-allocation-service:latest
```

## Troubleshooting

### Database Connection Issues

**Error: "failed SASL auth" or "connection refused"**

1. Verify container is running: `docker ps`
2. Check `.env` configuration:
   - `DB_HOST=127.0.0.1` (not `localhost`)
   - `DB_PORT=54321` (local port, not 5432)
   - `DB_PASSWORD=allocation_pass`
3. Verify port availability: `netstat -an | findstr 54321`
4. Check port mapping: `docker port allocation_postgres`

### Migration Issues

**Error: "migrate: no change"**

Migrations already applied. To revert:

```bash
make migrate-down
make migrate-up
```

### Recreate Database

```bash
docker-compose down -v
docker-compose up -d postgres
# Wait 5-10 seconds
make migrate-up
```

### Logs

**View service logs:**
```bash
# Development
make run

# Docker
docker logs -f allocation_postgres
```

**Log levels:**
- `debug`: Detailed debugging information
- `info`: General informational messages
- `warn`: Warning messages
- `error`: Error messages

## Additional Documentation

- [Architecture Documentation](./ARCHITECTURE.md) - Detailed architecture and design decisions
- [OpenAPI Specification](./api/openapi.yaml) - Complete API documentation
- [Database Migrations](./migrations/) - SQL migration files

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

© 2025 Denzell Griffith - [GPLv3](../LICENSE)

## Links

- [Go Documentation](https://golang.org/doc/)
- [Chi Router](https://github.com/go-chi/chi)
- [PostgreSQL Documentation](https://www.postgresql.org/docs/)
- [sqlc Documentation](https://docs.sqlc.dev/)
- [OpenAPI Specification](https://swagger.io/specification/)
