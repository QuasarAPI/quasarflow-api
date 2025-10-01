# ================================
# STAGE 1 - Build
# ================================
FROM golang:1.24 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o quasarflow-api ./cmd/api

# ================================
# STAGE 2 - Runtime
# ================================
FROM gcr.io/distroless/base-debian12 AS final

WORKDIR /app

COPY --from=builder /app/quasarflow-api .

EXPOSE 8080

ENTRYPOINT ["/app/quasarflow-api"]
