# QuasarFlow API

> REST API in Go for Stellar wallet management with AES-256-GCM encryption

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Architecture](https://img.shields.io/badge/Architecture-Clean%20Architecture-brightgreen)](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)

## 📋 Table of Contents

- [About](#-about)
- [Architecture](#-architecture)
- [Technologies](#-technologies)
- [Prerequisites](#-prerequisites)
- [Installation](#-installation)
- [Configuration](#-configuration)
- [Usage](#-usage)
- [API Endpoints](#-api-endpoints)
- [Migrations](#-migrations)
- [Docker](#-docker)
- [Project Structure](#-project-structure)

## 🎯 About

QuasarFlow API is a backend application developed in Go that provides a secure interface for creating and managing Stellar (blockchain) wallets. The API implements:

- ✅ **Stellar wallet generation** (testnet and mainnet)
- ✅ **AES-256-GCM encryption** for private keys
- ✅ **Balance queries** via Horizon API
- ✅ **Paginated wallet listing**
- ✅ **Health checks** for monitoring
- ✅ **Graceful shutdown** with configurable timeouts

## 🏗️ Architecture

The project follows **Clean Architecture** principles, organized in layers:

```
┌─────────────────────────────────────────┐
│          Interface Layer                │
│  (HTTP Handlers, Middlewares, Router)  │
└─────────────────┬───────────────────────┘
                  │
┌─────────────────▼───────────────────────┐
│          Use Case Layer                 │
│  (Business Logic, Application Rules)   │
└─────────────────┬───────────────────────┘
                  │
┌─────────────────▼───────────────────────┐
│         Domain Layer                    │
│  (Entities, Repository Interfaces)     │
└─────────────────┬───────────────────────┘
                  │
┌─────────────────▼───────────────────────┐
│      Infrastructure Layer               │
│  (Database, Stellar Client, Crypto)    │
└─────────────────────────────────────────┘
```

**Benefits:**
- ✅ Testability: Decoupled layers facilitate unit testing
- ✅ Maintainability: Clear separation of concerns
- ✅ Flexibility: Easy to swap implementations (e.g., change database)

## 🛠️ Technologies

- **[Go 1.21+](https://go.dev/)** - Programming language
- **[Gorilla Mux](https://github.com/gorilla/mux)** - HTTP router and URL matcher
- **[PostgreSQL](https://www.postgresql.org/)** - Relational database
- **[Stellar Go SDK](https://github.com/stellar/go)** - Stellar blockchain integration
- **[Uber Zap](https://github.com/uber-go/zap)** - Structured logging
- **[godotenv](https://github.com/joho/godotenv)** - Environment variable management

## 📦 Prerequisites

- **Go 1.21+** ([Installation](https://go.dev/doc/install))
- **PostgreSQL 12+** ([Installation](https://www.postgresql.org/download/))
- **Docker** (optional, for containerization)

## 🚀 Installation

### 1. Clone the repository

```bash
git clone https://github.com/QuasarAPI/quasarflow-api.git
cd quasarflow-api
```

### 2. Install dependencies

```bash
go mod download
```

### 3. Setup the database

Create a PostgreSQL database:

```bash
createdb quasarflow
```

Run migrations:

```bash
psql -U postgres -d quasarflow -f migrations/001_create_wallets_table.up.sql
```

### 4. Configure environment variables

Copy the example file and adjust settings:

```bash
cp .env.example .env
```

### 5. Build and run

```bash
go build -o quasarflow-api ./cmd/api
./quasarflow-api
```

Or directly with `go run`:

```bash
go run ./cmd/api/main.go
```

## ⚙️ Configuration

Edit the `.env` file with your settings:

```env
# Server
PORT=:8080
ENV=development

# Database
DATABASE_URL=postgresql://user:password@localhost:5432/quasarflow?sslmode=disable

# Stellar
STELLAR_HORIZON_URL=https://horizon-testnet.stellar.org
STELLAR_NETWORK=testnet

# Security (⚠️ IMPORTANT: Use 32-byte key for production)
ENCRYPTION_KEY=your-32-byte-encryption-key!!

# Logging
LOG_LEVEL=info
```

### Environment Variables

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `PORT` | HTTP server port | `:8080` | No |
| `ENV` | Environment (development/staging/production) | `development` | No |
| `DATABASE_URL` | PostgreSQL connection string | - | **Yes** |
| `STELLAR_HORIZON_URL` | Horizon API URL | `https://horizon-testnet.stellar.org` | No |
| `STELLAR_NETWORK` | Stellar network (testnet/mainnet) | `testnet` | No |
| `ENCRYPTION_KEY` | AES-256 key (32 bytes) | - | **Yes** |
| `LOG_LEVEL` | Log level (debug/info/warn/error) | `info` | No |

## 📡 Usage

### Create a Wallet

```bash
curl -X POST http://localhost:8080/api/v1/wallets \
  -H "Content-Type: application/json" \
  -d '{"network": "testnet"}'
```

**Response:**
```json
{
  "success": true,
  "data": {
    "id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
    "public_key": "GXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX",
    "network": "testnet",
    "created_at": "2025-10-01T12:34:56Z"
  }
}
```

### Get Wallet by ID

```bash
curl http://localhost:8080/api/v1/wallets/{wallet_id}
```

### Get Wallet Balances

```bash
curl http://localhost:8080/api/v1/wallets/{wallet_id}/balance
```

### List Wallets (with pagination)

```bash
curl "http://localhost:8080/api/v1/wallets?limit=10&offset=0"
```

### Health Check

```bash
curl http://localhost:8080/health
```

**Response:**
```json
{
  "success": true,
  "data": {
    "status": "healthy",
    "database": "healthy"
  }
}
```

## 🔌 API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/health` | Check application health |
| `POST` | `/api/v1/wallets` | Create new wallet |
| `GET` | `/api/v1/wallets` | List wallets (paginated) |
| `GET` | `/api/v1/wallets/{id}` | Get wallet by ID |
| `GET` | `/api/v1/wallets/{id}/balance` | Query blockchain balances |

## 🗄️ Migrations

### Apply migrations

```bash
psql -U postgres -d quasarflow -f migrations/001_create_wallets_table.up.sql
```

### Rollback migrations

```bash
psql -U postgres -d quasarflow -f migrations/001_create_wallets_table.down.sql
```

## 🐳 Docker

### Build image

```bash
docker build -t quasarflow-api .
```

### Run container

```bash
docker run -d \
  --name quasarflow \
  -p 8080:8080 \
  --env-file .env \
  quasarflow-api
```

### Docker Compose (coming soon)

```yaml
# TODO: Add docker-compose.yml with PostgreSQL + API
```

## 📁 Project Structure

```
quasarflow-api/
├── cmd/
│   └── api/
│       └── main.go                 # Application entry point
├── internal/
│   ├── config/
│   │   └── config.go               # Application configuration
│   ├── domain/
│   │   ├── stellar/
│   │   │   └── balance.go          # Balance entity (Stellar)
│   │   └── wallet/
│   │       ├── entity.go           # Wallet entity
│   │       └── repository.go       # Repository interface
│   ├── infrastructure/
│   │   ├── crypto/
│   │   │   └── encryption.go       # AES-256-GCM implementation
│   │   ├── database/
│   │   │   └── wallet_repository.go # PostgreSQL repository
│   │   └── stellar/
│   │       └── client.go           # Horizon API client
│   ├── interface/
│   │   └── http/
│   │       ├── handler/
│   │       │   ├── health_handler.go
│   │       │   └── wallet_handler.go
│   │       ├── middleware/
│   │       │   ├── cors.go
│   │       │   ├── logger.go
│   │       │   ├── recovery.go
│   │       │   └── request_id.go
│   │       ├── response/
│   │       │   └── response.go     # JSON response formatting
│   │       └── router.go           # Route definitions
│   └── usecase/
│       └── wallet/
│           ├── create_wallet.go
│           ├── get_balance.go
│           ├── get_wallet.go
│           └── list_wallets.go
├── pkg/
│   ├── errors/
│   │   ├── domain_errors.go        # Domain errors
│   │   └── error.go                # Custom error types
│   └── logger/
│       └── logger.go                # Logging interface
├── migrations/
│   ├── 001_create_wallets_table.up.sql
│   └── 001_create_wallets_table.down.sql
├── .env.example                     # Environment variables example
├── Dockerfile
├── go.mod
├── go.sum
├── LICENSE
└── README.md
```

## 🔒 Security

### Private Key Encryption

- **Algorithm:** AES-256-GCM (Galois/Counter Mode)
- **Key:** 32 bytes (256 bits)
- **Nonce:** Randomly generated for each operation
- **Encoding:** Base64 for storage

⚠️ **IMPORTANT:** Never commit `ENCRYPTION_KEY` to public repositories!

### Secure Key Generation

```bash
# Generate 32-byte (256-bit) key
openssl rand -base64 32
```

## 📝 TODO

- [ ] Add unit tests
- [ ] Add integration tests
- [ ] Implement JWT authentication
- [ ] Add rate limiting
- [ ] Create docker-compose.yml
- [ ] Add CI/CD (GitHub Actions)
- [ ] Implement Prometheus metrics
- [ ] Add Swagger/OpenAPI documentation
- [ ] Implement Stellar transactions
- [ ] Add multi-signature support

## 🤝 Contributing

Contributions are welcome! Feel free to open issues and pull requests.

1. Fork the project
2. Create your feature branch (`git checkout -b feature/MyFeature`)
3. Commit your changes (`git commit -m 'Add: new feature'`)
4. Push to the branch (`git push origin feature/MyFeature`)
5. Open a Pull Request

## 📄 License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

## 👨‍💻 Author

**QuasarAPI Team**

---

⭐ If this project was helpful to you, consider giving it a star on the repository!

