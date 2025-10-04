# Multi-stage Dockerfile for QuasarFlow API
# Production-ready with security best practices

# ================================
# STAGE 1 - Build
# ================================
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Create non-root user for build
RUN adduser -D -s /bin/sh -u 10001 appuser

WORKDIR /app

# Copy go mod files first for better caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Build the application with optimizations
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -a -installsuffix cgo \
    -ldflags='-w -s -extldflags "-static"' \
    -o quasarflow-api ./cmd/api

# ================================
# STAGE 2 - Runtime
# ================================
FROM alpine:3.19 AS final

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata curl && \
    update-ca-certificates

# Create non-root user
RUN adduser -D -s /bin/sh -u 10001 appuser

# Create app directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/quasarflow-api .
COPY --from=builder /app/migrations ./migrations

# Set proper ownership
RUN chown -R appuser:appuser /app

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8080

# Add health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

# Run the application
CMD ["./quasarflow-api"]
