# ================================
# STAGE 1 - Build
# ================================
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod ./
COPY go.sum* ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o quasarflow-api ./cmd/api

# ================================
# STAGE 2 - Runtime
# ================================
FROM gcr.io/distroless/base-debian12 AS final

WORKDIR /app

COPY --from=builder /app/quasarflow-api .

EXPOSE 8080

ENTRYPOINT ["/app/quasarflow-api"]
