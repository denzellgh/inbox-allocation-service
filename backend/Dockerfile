# ============================================
# Stage 1: Builder
# ============================================
ARG GO_VERSION=1.25.2

FROM golang:${GO_VERSION}-alpine AS builder

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

# Copiar dependencias primero (cache de layers)
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Copiar código fuente
COPY . .

# Compilar binario estático optimizado
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o /app/main \
    ./cmd/server

# ============================================
# Stage 2: Runtime (imagen mínima ~10MB)
# ============================================
FROM alpine:3.20

# Certificados SSL y timezone data
RUN apk --no-cache add ca-certificates tzdata \
    && rm -rf /var/cache/apk/*

# Usuario no-root para seguridad
RUN addgroup -g 1000 appgroup \
    && adduser -u 1000 -G appgroup -s /bin/sh -D appuser

WORKDIR /app

# Copiar binario desde builder
COPY --from=builder /app/main .

# Cambiar ownership y permisos
RUN chown -R appuser:appgroup /app

USER appuser

# Puerto expuesto (debe coincidir con ECS task definition)
EXPOSE 8080

# Healthcheck para ECS
HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Ejecutar binario
ENTRYPOINT ["./main"]