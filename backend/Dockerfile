# ============================================
# Inbox Allocation Service - Multi-Stage Build
# ============================================

# ============================================
# Stage 1: Build
# ============================================
ARG GO_VERSION=1.25.2

FROM golang:${GO_VERSION}-alpine AS builder

# Build arguments
ARG VERSION=dev
ARG BUILD_TIME=unknown
ARG GIT_COMMIT=unknown

# Install build dependencies
RUN apk add --no-cache \
    git \
    ca-certificates \
    tzdata

# Set working directory
WORKDIR /build

# Copy go mod files first (better layer caching)
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Build the application with optimizations
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s \
        -X main.Version=${VERSION} \
        -X main.BuildTime=${BUILD_TIME} \
        -X main.GitCommit=${GIT_COMMIT}" \
    -o /build/bin/server \
    ./cmd/server

# ============================================
# Stage 2: Production
# ============================================
FROM alpine:3.19 AS production

# Labels
LABEL maintainer="your-team@example.com"
LABEL org.opencontainers.image.title="Inbox Allocation Service"
LABEL org.opencontainers.image.description="High-performance microservice for inbox management and conversation allocation"
LABEL org.opencontainers.image.source="https://github.com/your-org/inbox-allocation-service"

# Install runtime dependencies
RUN apk add --no-cache \
    ca-certificates \
    tzdata \
    wget

# Create non-root user
RUN addgroup -g 1000 -S appgroup && \
    adduser -u 1000 -S appuser -G appgroup

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/bin/server /app/server

# Copy migrations for potential startup migrations
COPY --from=builder /build/migrations /app/migrations

# Copy OpenAPI spec for documentation
COPY --from=builder /build/api /app/api

# Set ownership
RUN chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser

# Expose HTTP port
EXPOSE 8080

# Environment defaults
ENV SERVER_HOST=0.0.0.0 \
    SERVER_PORT=8080 \
    LOG_LEVEL=info \
    LOG_FORMAT=json

# Health check
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the application
ENTRYPOINT ["/app/server"]

# ============================================
# Stage 3: Development (optional)
# ============================================
FROM golang:${GO_VERSION}-alpine AS development

# Install development tools
RUN apk add --no-cache \
    git \
    make \
    bash

# Install Go tools
RUN go install github.com/air-verse/air@latest && \
    go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest && \
    go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Default command for development
CMD ["air", "-c", ".air.toml"]
