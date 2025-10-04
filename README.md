# QuasarFlow API

> Blockchain abstraction API that simplifies Stellar network integration through REST endpoints

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

## About QuasarFlow

Building blockchain applications typically requires understanding cryptographic key management, learning blockchain-specific SDKs, and handling complex network scenarios. **QuasarFlow API** eliminates this complexity by providing a simple REST interface to the Stellar network.

Instead of learning blockchain protocols, developers can build blockchain-powered applications using familiar REST API calls:

```bash
# Create a blockchain wallet
POST /api/v1/wallets

# Check wallet balance  
GET /api/v1/wallets/{id}/balance

# Send payments
POST /api/v1/wallets/{id}/payment
```

**That's it.** No blockchain knowledge required.

## Key Features

**For Developers:**
- **Simple REST API** - No blockchain SDK learning curve
- **Enterprise Security** - JWT authentication, rate limiting, and comprehensive security headers
- **Secure by default** - AES-256-GCM encryption for private keys
- **External Wallet Support** - Verify ownership of existing Stellar wallets using SEP-10 standards
- **Production ready** - Built with enterprise-grade architecture and security
- **Well documented** - Clear API specifications and examples

**For Businesses:**
- **Faster time to market** - Build blockchain features in days, not months
- **Reduced development cost** - No need for specialized blockchain developers
- **Lower maintenance** - We handle blockchain complexity and updates
- **Scalable infrastructure** - Built to handle enterprise workloads
- **Security compliance** - Enterprise-grade security features built-in

## Use Cases

**Fintech Applications**
- Add crypto wallet functionality to existing banking apps
- Enable cryptocurrency payments in e-commerce platforms

**Gaming & Digital Assets**  
- Manage in-game assets on blockchain
- Create NFT marketplaces and digital collectibles

**Enterprise Solutions**
- Add blockchain capabilities to business systems
- Enable cross-border payments and remittances

## Technology Stack

- **[Go 1.21+](https://go.dev/)** - High-performance backend language
- **[PostgreSQL](https://www.postgresql.org/)** - Reliable data persistence
- **[Stellar Network](https://stellar.org/)** - Fast, low-cost blockchain
- **[Clean Architecture](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)** - Maintainable, testable codebase


## Quick Start

### Docker (Recommended)

```bash
# Clone and setup
git clone https://github.com/QuasarAPI/quasarflow-api.git
cd quasarflow-api

# Setup environment
cp .env.example .env
nano .env  # customize settings

# Start all services
docker-compose up -d

# View logs
docker-compose logs -f quasarflow-api
```

### Manual Setup

```bash
# Prerequisites: Go 1.21+, PostgreSQL 12+
git clone https://github.com/QuasarAPI/quasarflow-api.git
cd quasarflow-api
go mod download

# Setup database
createdb quasarflow
./scripts/db-manage.sh migrate-up

# Configure and run
cp .env.example .env
go run ./cmd/api/main.go
```

## Quick Example

```bash
# 1. Login and get token
TOKEN=$(curl -s -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username": "admin", "password": "admin123"}' | jq -r '.data.token')

# 2. Create a wallet
WALLET_ID=$(curl -s -X POST http://localhost:8080/api/v1/wallets \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"network": "local"}' | jq -r '.data.id')

# 3. Fund wallet (development only)
curl -X POST http://localhost:8080/api/v1/wallets/$WALLET_ID/fund \
  -H "Authorization: Bearer $TOKEN"

# 4. Check balance
curl http://localhost:8080/api/v1/wallets/$WALLET_ID/balance \
  -H "Authorization: Bearer $TOKEN"
```

## External Wallet Support

QuasarFlow API supports users with existing Stellar wallets through **ownership verification**. Users can prove they own a wallet without registering with the platform.

### Quick Example - Verify Existing Wallet

```bash
# 1. Generate challenge for wallet
CHALLENGE=$(curl -s http://localhost:8080/api/v1/accounts/GABC123.../challenge | jq -r '.challenge')

# 2. Sign challenge with private key (client-side)
# This requires a Stellar SDK implementation in your application
SIGNATURE="base64_encoded_signature"

# 3. Verify ownership
curl -X POST http://localhost:8080/api/v1/accounts/GABC123.../verify-ownership \
  -H "Content-Type: application/json" \
  -d '{"signature": "'$SIGNATURE'", "message": "'$CHALLENGE'"}'

# 4. Alternative: Verify via recent transaction
curl -X POST http://localhost:8080/api/v1/accounts/GABC123.../verify-transaction \
  -H "Content-Type: application/json" \
  -d '{"transaction_hash": "transaction_hash_here"}'
```

### Supported Verification Methods

- **🔐 Message Signing (SEP-10)** - Cryptographically sign a challenge
- **📝 Transaction Proof** - Use recent signed transactions as proof
- **📊 Account Activity** - Verify based on account existence and activity

For complete implementation examples in JavaScript, Go, and Python, see [Wallet Ownership Verification Guide](docs/WALLET_OWNERSHIP_VERIFICATION.md).

**Demo Credentials:** `admin/admin123` or `user/user123`

## Configuration

### Environment Setup

```bash
# Copy and customize environment file
cp .env.example .env
nano .env  # Edit with your settings
```

### Key Variables

| Variable | Description | Example |
|----------|-------------|---------|
| `ENV` | Environment | `development`, `production` |
| `STELLAR_NETWORK` | Stellar network | `local`, `testnet`, `mainnet` |
| `ENCRYPTION_KEY` | AES key (32 bytes) | `openssl rand -base64 32` |
| `JWT_SECRET` | JWT secret (32+ chars) | `SecureProductionSecret123!` |
| `ALLOWED_ORIGINS` | CORS origins | `https://yourdomain.com` |

### Network Modes

- **Local**: Docker Stellar network + Friendbot for testing
- **Testnet**: Stellar testnet for staging  
- **Mainnet**: Production Stellar network

## API Overview

The API provides REST endpoints for blockchain operations with JWT authentication:

- **Authentication**: `/auth/login`, `/auth/logout`
- **Wallet Management**: `/api/v1/wallets`
- **Transactions**: Payments, balance queries, history
- **Health Check**: `/health`

> **📚 Complete API documentation:** See `docs/API.md` for detailed endpoint specifications

## Architecture

QuasarFlow API follows Clean Architecture principles for maintainability and testability:

```
┌─────────────────────────────────────────┐
│             HTTP Layer                  │
│    (REST API, Middleware, Routing)     │
└─────────────────┬───────────────────────┘
                  │
┌─────────────────▼───────────────────────┐
│           Use Case Layer                │
│      (Business Logic & Rules)          │
└─────────────────┬───────────────────────┘
                  │
┌─────────────────▼───────────────────────┐
│           Domain Layer                  │
│     (Entities & Interfaces)            │
└─────────────────┬───────────────────────┘
                  │
┌─────────────────▼───────────────────────┐
│        Infrastructure Layer             │
│  (Database, Stellar Client, Crypto)    │
└─────────────────────────────────────────┘
```

This architecture ensures:
- **Testability** - Easy to unit test business logic
- **Flexibility** - Simple to swap implementations
- **Maintainability** - Clear separation of concerns

## Security

### Enterprise Security Features

- **JWT Authentication** with configurable expiration
- **Rate Limiting** to prevent abuse
- **CORS Protection** with configurable origins
- **Security Headers** (XSS, clickjacking, MIME protection)
- **AES-256-GCM Encryption** for private key storage
- **Input Validation** and sanitization

### Authentication Flow

1. Login with credentials → Receive JWT token
2. Include token in `Authorization: Bearer <token>` header
3. Access protected endpoints with valid token

## Docker Development

### Available Services

- **API**: `http://localhost:8080`
- **PostgreSQL**: `localhost:5432`
- **Stellar Horizon**: `http://localhost:8000` (local network)
- **Friendbot**: `http://localhost:8000/friendbot` (funding)

### Management Commands

```bash
# Database migrations
docker-compose run --rm migrate

# API testing
./scripts/test-api.sh

# Service management
docker-compose ps
docker-compose restart quasarflow-api
docker-compose logs -f quasarflow-api
```

## Project Structure

```
quasarflow-api/
├── cmd/api/                    # Application entry point
├── internal/
│   ├── config/                 # Configuration management
│   ├── domain/                 # Business entities & interfaces
│   ├── infrastructure/         # Database, Stellar, Crypto
│   ├── interface/http/         # REST API, Handlers, Middleware
│   └── usecase/               # Business logic
├── migrations/                 # Database schema
├── pkg/                       # Shared utilities
└── docs/                      # API documentation
```

## Roadmap

**Current Features:**
- ✅ Wallet creation and management
- ✅ Secure private key storage (AES-256-GCM)
- ✅ Payment transactions (XLM and custom assets)
- ✅ Multi-network support (local, testnet, mainnet)
- ✅ JWT Authentication & Enterprise Security
- ✅ Docker development environment

**Roadmap:**
- Multi-signature wallet support
- Webhook notification system
- Database user management
- Advanced role-based permissions
- Support for additional blockchains
- Advanced transaction types (escrow, atomic swaps)
- Analytics and reporting dashboard
- SDKs for popular languages

## Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/amazing-feature`
3. Make your changes and add tests
4. Commit your changes: `git commit -m 'Add amazing feature'`
5. Push to the branch: `git push origin feature/amazing-feature`
6. Open a Pull Request

## Documentation

- **[API Documentation](docs/API.md)**: Complete endpoint specifications
- **[Implementation Guide](IMPLEMENTATION_SUMMARY.md)**: Technical architecture details
- **Quick Troubleshooting**: `curl http://localhost:8080/health`

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

**Quasar Flow** - Making blockchain development simple and accessible for everyone.