#!/bin/bash

# QuasarFlow API - Simplified Environment Setup Script
# Sets up QuasarFlow API using .env configuration

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Function to print colored output
print_header() {
    echo ""
    echo -e "${CYAN}========================================${NC}"
    echo -e "${CYAN} $1 ${NC}"
    echo -e "${CYAN}========================================${NC}"
    echo ""
}

print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Show usage information
show_usage() {
    echo ""
    echo "QuasarFlow API - Environment Setup"
    echo ""
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  --clean              Clean all data before setup"
    echo "  --no-build           Skip building containers"
    echo "  --verbose            Show verbose output"
    echo "  --help, -h           Show this help message"
    echo ""
    echo "Environment Configuration:"
    echo "  1. Copy .env.example to .env"
    echo "  2. Customize .env for your environment"
    echo "  3. Run this setup script"
    echo ""
    echo "Examples:"
    echo "  cp .env.example .env && $0          # Basic setup"
    echo "  $0 --clean                         # Clean setup"
    echo "  $0 --no-build                      # Skip build"
    echo ""
}

# Parse command line arguments
CLEAN_DATA=false
NO_BUILD=false
VERBOSE=false

while [[ $# -gt 0 ]]; do
    case $1 in
        --clean)
            CLEAN_DATA=true
            shift
            ;;
        --no-build)
            NO_BUILD=true
            shift
            ;;
        --verbose)
            VERBOSE=true
            shift
            ;;
        --help|-h)
            show_usage
            exit 0
            ;;
        *)
            print_error "Unknown option: $1"
            show_usage
            exit 1
            ;;
    esac
done

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    print_error "Docker is not running. Please start Docker and try again."
    exit 1
fi

# Check if docker-compose is available
if ! command -v docker-compose &> /dev/null; then
    print_error "docker-compose is not installed. Please install it and try again."
    exit 1
fi

# Check if .env file exists
if [[ ! -f ".env" ]]; then
    print_error ".env file not found!"
    echo ""
    print_status "To create .env file:"
    print_status "  1. Copy the example: cp .env.example .env"
    print_status "  2. Edit .env with your settings"
    print_status "  3. Run this script again"
    echo ""
    if [[ -f ".env.example" ]]; then
        print_status "Found .env.example - would you like to copy it to .env? (y/N)"
        read -r -n 1 response
        echo ""
        if [[ $response =~ ^[Yy]$ ]]; then
            cp .env.example .env
            print_success ".env file created from .env.example"
            print_warning "Please review and customize .env before continuing"
            print_status "Edit .env with: nano .env"
            exit 0
        fi
    fi
    exit 1
fi

# Load environment variables
print_status "Loading environment configuration..."
set -a
source .env
set +a

print_header "QuasarFlow API Setup"
print_status "Environment: ${ENV:-development}"
print_status "Project: ${COMPOSE_PROJECT_NAME:-quasarflow}"
print_status "Database: ${POSTGRES_DB:-quasarflow}"
print_status "Stellar Network: ${STELLAR_NETWORK:-local}"
print_status "API Port: ${API_PORT:-8080}"

# Validation
if [[ -z "$ENCRYPTION_KEY" ]]; then
    print_error "ENCRYPTION_KEY is not set in .env file"
    print_error "Generate one with: openssl rand -base64 32"
    exit 1
fi

if [[ ${#ENCRYPTION_KEY} -ne 32 ]]; then
    print_warning "ENCRYPTION_KEY should be exactly 32 characters"
    print_warning "Current length: ${#ENCRYPTION_KEY}"
fi

# Production warnings
if [[ "${ENV}" == "production" ]]; then
    print_header "PRODUCTION ENVIRONMENT DETECTED"
    print_warning "Please ensure you have:"
    print_warning "  ✓ Changed default passwords"
    print_warning "  ✓ Set secure encryption key"
    print_warning "  ✓ Configured SSL certificates"
    print_warning "  ✓ Set up monitoring and backups"
    print_warning "  ✓ Reviewed security settings"
    echo ""
    read -p "Continue with production setup? (y/N): " -n 1 -r
    echo ""
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        print_status "Setup cancelled"
        exit 0
    fi
fi

# Clean data if requested
if [[ "$CLEAN_DATA" == "true" ]]; then
    print_warning "Cleaning all data (volumes will be removed)..."
    docker-compose down -v --remove-orphans
    docker system prune -f --volumes
    print_success "Data cleaned"
fi

# Stop existing containers
print_status "Stopping existing containers..."
docker-compose down --remove-orphans

# Build containers unless --no-build is specified
if [[ "$NO_BUILD" != "true" ]]; then
    print_status "Building containers..."
    if [[ "$VERBOSE" == "true" ]]; then
        docker-compose build --no-cache
    else
        docker-compose build --no-cache > /dev/null 2>&1
    fi
    print_success "Containers built successfully"
else
    print_status "Skipping container build"
fi

# Start PostgreSQL
print_status "Starting PostgreSQL database..."
docker-compose up -d postgres

# Wait for PostgreSQL to be ready
print_status "Waiting for PostgreSQL to be ready..."
POSTGRES_READY=false
RETRY_COUNT=0
MAX_RETRIES=30

while [ "$POSTGRES_READY" = false ] && [ $RETRY_COUNT -lt $MAX_RETRIES ]; do
    if docker-compose exec -T postgres pg_isready -U "${POSTGRES_USER:-stellar}" -d "${POSTGRES_DB:-quasarflow}" > /dev/null 2>&1; then
        POSTGRES_READY=true
        print_success "PostgreSQL is ready!"
    else
        echo -n "."
        sleep 2
        RETRY_COUNT=$((RETRY_COUNT + 1))
    fi
done

if [ "$POSTGRES_READY" = false ]; then
    print_error "PostgreSQL failed to start. Check logs:"
    print_error "  docker-compose logs postgres"
    exit 1
fi

# Run database migrations
print_status "Running database migrations..."
docker-compose run --rm migrate
print_success "Database migrations completed"

# Start Stellar network (not needed for mainnet)
if [[ "${STELLAR_NETWORK:-local}" != "mainnet" ]]; then
    print_status "Starting Stellar network (${STELLAR_NETWORK:-local})..."
    docker-compose up -d stellar-quickstart

    print_status "Waiting for Stellar network to be ready..."
    STELLAR_READY=false
    RETRY_COUNT=0
    MAX_RETRIES=40

    while [ "$STELLAR_READY" = false ] && [ $RETRY_COUNT -lt $MAX_RETRIES ]; do
        if curl -f "http://localhost:${STELLAR_HORIZON_PORT:-8000}/" > /dev/null 2>&1; then
            STELLAR_READY=true
            print_success "Stellar network is ready!"
        else
            echo -n "."
            sleep 10
            RETRY_COUNT=$((RETRY_COUNT + 1))
        fi
    done

    if [ "$STELLAR_READY" = false ]; then
        print_error "Stellar network failed to start. Check logs:"
        print_error "  docker-compose logs stellar-quickstart"
        exit 1
    fi
else
    print_status "Using external Stellar mainnet - no local network needed"
fi

# Start QuasarFlow API
print_status "Starting QuasarFlow API..."
docker-compose up -d quasarflow-api

# Wait for API to be ready
print_status "Waiting for API to be ready..."
API_READY=false
RETRY_COUNT=0
MAX_RETRIES=20

while [ "$API_READY" = false ] && [ $RETRY_COUNT -lt $MAX_RETRIES ]; do
    if curl -f "http://localhost:${API_PORT:-8080}/health" > /dev/null 2>&1; then
        API_READY=true
        print_success "QuasarFlow API is ready!"
    else
        echo -n "."
        sleep 3
        RETRY_COUNT=$((RETRY_COUNT + 1))
    fi
done

if [ "$API_READY" = false ]; then
    print_warning "API might still be starting. Check logs:"
    print_warning "  docker-compose logs quasarflow-api"
fi

# Show final status
print_header "Setup Complete! 🎉"
print_status "Environment: ${ENV:-development}"
print_status "Project: ${COMPOSE_PROJECT_NAME:-quasarflow}"
echo ""

print_status "📋 Available services:"
echo "  • QuasarFlow API:     http://localhost:${API_PORT:-8080}"
echo "  • API Health Check:   http://localhost:${API_PORT:-8080}/health"

if [[ "${STELLAR_NETWORK:-local}" != "mainnet" ]]; then
    echo "  • Stellar Horizon:    http://localhost:${STELLAR_HORIZON_PORT:-8000}"
    echo "  • Stellar Lab:        http://localhost:${STELLAR_HORIZON_PORT:-8000}/lab"
    if [[ "${STELLAR_NETWORK:-local}" == "local" ]]; then
        echo "  • Friendbot:          http://localhost:${STELLAR_HORIZON_PORT:-8000}/friendbot"
    fi
fi

echo "  • PostgreSQL:         localhost:${POSTGRES_PORT:-5432} (${POSTGRES_USER:-stellar}/${POSTGRES_PASSWORD:-stellar123})"
echo ""

print_status "📖 Management commands:"
echo "  • View logs:          docker-compose logs -f [service]"
echo "  • Stop all:           docker-compose down"
echo "  • Restart API:        docker-compose restart quasarflow-api"
echo "  • Database tools:     ./scripts/db-manage.sh [command] --docker"
echo ""

# Environment-specific instructions
case "${ENV:-development}" in
    development)
        print_status "🧪 Development quick start:"
        echo "  1. Create wallet:     curl -X POST http://localhost:${API_PORT:-8080}/api/v1/wallets -H 'Content-Type: application/json' -d '{\"network\":\"${STELLAR_NETWORK:-local}\"}'"
        echo "  2. List wallets:      curl http://localhost:${API_PORT:-8080}/api/v1/wallets"
        if [[ "${STELLAR_NETWORK:-local}" == "local" ]]; then
            echo "  3. Fund wallet:       curl -X POST http://localhost:${API_PORT:-8080}/api/v1/wallets/WALLET_ID/fund"
        fi
        echo ""
        print_status "🔧 Development tools:"
        echo "  • API Testing:        ./scripts/test-api.sh"
        echo "  • Database Reset:     ./scripts/db-manage.sh migrate-reset --docker"
        ;;

    staging)
        print_status "🎭 Staging environment ready:"
        echo "  • Network: Stellar ${STELLAR_NETWORK:-testnet}"
        echo "  • Database: ${POSTGRES_DB:-quasarflow}"
        echo "  • Testing: ./scripts/test-api.sh --network ${STELLAR_NETWORK:-testnet}"
        ;;

    production)
        print_status "🚀 Production environment ready:"
        echo "  • Network: Stellar ${STELLAR_NETWORK:-mainnet}"
        echo "  • Database: ${POSTGRES_DB:-quasarflow}"
        echo ""
        print_status "⚠️  Critical production monitoring:"
        echo "  1. Health check:      curl http://localhost:${API_PORT:-8080}/health"
        echo "  2. Monitor logs:      docker-compose logs -f --tail=100"
        echo "  3. Database backup:   ./scripts/db-manage.sh backup --docker"
        echo "  4. Container status:  docker-compose ps"
        print_warning "Ensure monitoring, SSL/TLS, and backup systems are active!"
        ;;
esac

echo ""
print_success "Setup completed successfully! 🚀"
echo ""
print_status "Container Status:"
docker-compose ps
