#!/bin/bash

# QuasarFlow API - Database Management Script
# Provides utilities for database operations, migrations, and testing

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default values
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_USER="${DB_USER:-stellar}"
DB_PASSWORD="${DB_PASSWORD:-stellar123}"
DB_NAME="${DB_NAME:-quasarflow}"
MIGRATIONS_DIR="${MIGRATIONS_DIR:-./migrations}"

# Function to print colored output
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
    echo "QuasarFlow API - Database Management"
    echo ""
    echo "Usage: $0 COMMAND [OPTIONS]"
    echo ""
    echo "Commands:"
    echo "  migrate-up          Apply all pending migrations"
    echo "  migrate-down        Rollback last migration"
    echo "  migrate-reset       Reset database (drop and recreate)"
    echo "  migrate-status      Show migration status"
    echo "  create-migration    Create new migration file"
    echo "  seed                Insert test data"
    echo "  backup              Create database backup"
    echo "  restore             Restore database from backup"
    echo "  test-data           Insert test wallets for development"
    echo "  clean               Clean test data"
    echo "  stats               Show database statistics"
    echo "  connect             Connect to database via psql"
    echo ""
    echo "Options:"
    echo "  --host HOST         Database host (default: $DB_HOST)"
    echo "  --port PORT         Database port (default: $DB_PORT)"
    echo "  --user USER         Database user (default: $DB_USER)"
    echo "  --password PASS     Database password (default: $DB_PASSWORD)"
    echo "  --database DB       Database name (default: $DB_NAME)"
    echo "  --docker            Use docker-compose database"
    echo ""
    echo "Examples:"
    echo "  $0 migrate-up"
    echo "  $0 migrate-up --docker"
    echo "  $0 create-migration add_user_table"
    echo "  $0 seed --docker"
    echo "  $0 backup > backup.sql"
    echo ""
}

# Parse command line arguments
COMMAND=""
USE_DOCKER=false

while [[ $# -gt 0 ]]; do
    case $1 in
        --host)
            DB_HOST="$2"
            shift 2
            ;;
        --port)
            DB_PORT="$2"
            shift 2
            ;;
        --user)
            DB_USER="$2"
            shift 2
            ;;
        --password)
            DB_PASSWORD="$2"
            shift 2
            ;;
        --database)
            DB_NAME="$2"
            shift 2
            ;;
        --docker)
            USE_DOCKER=true
            DB_HOST="localhost"
            shift
            ;;
        --help|-h)
            show_usage
            exit 0
            ;;
        *)
            if [[ -z "$COMMAND" ]]; then
                COMMAND="$1"
            fi
            shift
            ;;
    esac
done

# Validate command
if [[ -z "$COMMAND" ]]; then
    print_error "No command specified"
    show_usage
    exit 1
fi

# Database connection string
DB_URL="postgresql://$DB_USER:$DB_PASSWORD@$DB_HOST:$DB_PORT/$DB_NAME?sslmode=disable"

# Function to check if database is accessible
check_db_connection() {
    if $USE_DOCKER; then
        if ! docker-compose ps postgres | grep -q "Up"; then
            print_error "PostgreSQL container is not running. Start with: docker-compose up -d postgres"
            exit 1
        fi
    fi

    if ! PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "SELECT 1;" > /dev/null 2>&1; then
        print_error "Cannot connect to database: $DB_URL"
        print_error "Make sure PostgreSQL is running and credentials are correct"
        exit 1
    fi
}

# Function to run SQL file
run_sql_file() {
    local file=$1
    if [[ ! -f "$file" ]]; then
        print_error "SQL file not found: $file"
        exit 1
    fi

    print_status "Executing: $file"
    PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -f "$file"
}

# Function to run SQL command
run_sql() {
    local sql=$1
    PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "$sql"
}

# Create new migration file
create_migration() {
    local name=$1
    if [[ -z "$name" ]]; then
        print_error "Migration name is required"
        echo "Usage: $0 create-migration migration_name"
        exit 1
    fi

    # Generate timestamp
    local timestamp=$(date +%Y%m%d%H%M%S)
    local up_file="${MIGRATIONS_DIR}/${timestamp}_${name}.up.sql"
    local down_file="${MIGRATIONS_DIR}/${timestamp}_${name}.down.sql"

    # Create migrations directory if it doesn't exist
    mkdir -p "$MIGRATIONS_DIR"

    # Create up migration file
    cat > "$up_file" << EOF
-- Migration: ${name}
-- Created: $(date)
-- Description: Add your migration description here

BEGIN;

-- Add your migration SQL here
-- Example:
-- CREATE TABLE example (
--     id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
--     name VARCHAR(255) NOT NULL,
--     created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
-- );

COMMIT;
EOF

    # Create down migration file
    cat > "$down_file" << EOF
-- Rollback migration: ${name}
-- Created: $(date)

BEGIN;

-- Add your rollback SQL here
-- Example:
-- DROP TABLE IF EXISTS example;

COMMIT;
EOF

    print_success "Created migration files:"
    print_status "  Up:   $up_file"
    print_status "  Down: $down_file"
}

# Insert test data
insert_test_data() {
    print_status "Inserting test data..."

    local sql="
INSERT INTO wallets (id, public_key, encrypted_key, network, created_at, updated_at)
VALUES
    ('550e8400-e29b-41d4-a716-446655440001',
     'GAHK7EEG2WWHVKDNT4CEQFZGKF2LGDSW2IVM4S5DP42RBW3K6BTODB4A',
     'encrypted_key_placeholder_1',
     'local',
     NOW() - INTERVAL '1 day',
     NOW() - INTERVAL '1 day'),
    ('550e8400-e29b-41d4-a716-446655440002',
     'GBQMRS5EK6FCKAYJ3Z57VZ6U6HKAHOWWZJ5B6Q7BWZL5U5M3CLFYYLRA',
     'encrypted_key_placeholder_2',
     'testnet',
     NOW() - INTERVAL '2 hours',
     NOW() - INTERVAL '2 hours'),
    ('550e8400-e29b-41d4-a716-446655440003',
     'GCRNL5U5NWBGPB2JV4GLVNWK7WKJXPJV7NWBGPB2JV4GLVNWK7WKJX',
     'encrypted_key_placeholder_3',
     'local',
     NOW() - INTERVAL '30 minutes',
     NOW() - INTERVAL '30 minutes')
ON CONFLICT (public_key) DO NOTHING;"

    run_sql "$sql"
    print_success "Test data inserted successfully"
}

# Clean test data
clean_test_data() {
    print_warning "This will remove all test wallets. Continue? (y/N)"
    read -r confirm
    if [[ $confirm != [yY] ]]; then
        print_status "Operation cancelled"
        exit 0
    fi

    local sql="DELETE FROM wallets WHERE id IN (
        '550e8400-e29b-41d4-a716-446655440001',
        '550e8400-e29b-41d4-a716-446655440002',
        '550e8400-e29b-41d4-a716-446655440003'
    );"

    run_sql "$sql"
    print_success "Test data cleaned successfully"
}

# Show database statistics
show_stats() {
    print_status "Database Statistics"
    echo ""

    # Table sizes
    local sql="
SELECT
    schemaname as schema,
    tablename as table,
    pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) as size,
    pg_total_relation_size(schemaname||'.'||tablename) as size_bytes
FROM pg_tables
WHERE schemaname NOT IN ('information_schema', 'pg_catalog')
ORDER BY size_bytes DESC;
"

    echo "📊 Table Sizes:"
    run_sql "$sql"
    echo ""

    # Wallet statistics
    local wallet_stats="
SELECT
    network,
    COUNT(*) as wallet_count,
    MIN(created_at) as first_created,
    MAX(created_at) as last_created
FROM wallets
GROUP BY network
ORDER BY wallet_count DESC;
"

    echo "👛 Wallet Statistics:"
    run_sql "$wallet_stats"
    echo ""

    # Database size
    local db_size="
SELECT
    pg_database.datname as database_name,
    pg_size_pretty(pg_database_size(pg_database.datname)) as size
FROM pg_database
WHERE datname = '$DB_NAME';
"

    echo "💾 Database Size:"
    run_sql "$db_size"
}

# Create database backup
create_backup() {
    local backup_file="quasarflow_backup_$(date +%Y%m%d_%H%M%S).sql"
    print_status "Creating backup: $backup_file"

    PGPASSWORD=$DB_PASSWORD pg_dump -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME > "$backup_file"
    print_success "Backup created: $backup_file"
}

# Restore database from backup
restore_backup() {
    local backup_file=$1
    if [[ -z "$backup_file" ]]; then
        print_error "Backup file is required"
        echo "Usage: $0 restore backup_file.sql"
        exit 1
    fi

    if [[ ! -f "$backup_file" ]]; then
        print_error "Backup file not found: $backup_file"
        exit 1
    fi

    print_warning "This will replace all data in database '$DB_NAME'. Continue? (y/N)"
    read -r confirm
    if [[ $confirm != [yY] ]]; then
        print_status "Operation cancelled"
        exit 0
    fi

    print_status "Restoring from backup: $backup_file"
    PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME < "$backup_file"
    print_success "Database restored successfully"
}

# Connect to database
connect_db() {
    print_status "Connecting to database: $DB_NAME"
    PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME
}

# Execute commands
case $COMMAND in
    migrate-up)
        check_db_connection
        print_status "Applying migrations..."
        for file in $(ls ${MIGRATIONS_DIR}/*up.sql 2>/dev/null | sort); do
            run_sql_file "$file"
        done
        print_success "Migrations applied successfully"
        ;;

    migrate-down)
        check_db_connection
        print_status "Rolling back last migration..."
        # Find the most recent down migration
        local last_down=$(ls ${MIGRATIONS_DIR}/*down.sql 2>/dev/null | sort -r | head -n 1)
        if [[ -n "$last_down" ]]; then
            run_sql_file "$last_down"
            print_success "Migration rolled back successfully"
        else
            print_warning "No down migrations found"
        fi
        ;;

    migrate-reset)
        check_db_connection
        print_warning "This will DROP and recreate the database. Continue? (y/N)"
        read -r confirm
        if [[ $confirm == [yY] ]]; then
            print_status "Dropping database..."
            run_sql "DROP SCHEMA public CASCADE; CREATE SCHEMA public;"
            print_status "Recreating tables..."
            for file in $(ls ${MIGRATIONS_DIR}/*up.sql 2>/dev/null | sort); do
                run_sql_file "$file"
            done
            print_success "Database reset complete"
        else
            print_status "Operation cancelled"
        fi
        ;;

    migrate-status)
        check_db_connection
        print_status "Migration Status"
        echo ""
        echo "Available migrations:"
        ls ${MIGRATIONS_DIR}/*.sql 2>/dev/null | sort || echo "No migrations found"
        ;;

    create-migration)
        shift
        create_migration "$1"
        ;;

    seed|test-data)
        check_db_connection
        insert_test_data
        ;;

    clean)
        check_db_connection
        clean_test_data
        ;;

    backup)
        check_db_connection
        create_backup
        ;;

    restore)
        check_db_connection
        shift
        restore_backup "$1"
        ;;

    stats)
        check_db_connection
        show_stats
        ;;

    connect)
        check_db_connection
        connect_db
        ;;

    *)
        print_error "Unknown command: $COMMAND"
        show_usage
        exit 1
        ;;
esac
