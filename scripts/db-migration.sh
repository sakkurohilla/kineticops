#!/usr/bin/env bash
set -e

# Get the directory of the script
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Source the .env file in backend
source "$SCRIPT_DIR/../backend/.env"

# Define migrations directory
MIGRATIONS_DIR="$SCRIPT_DIR/../backend/migrations/postgres"

echo "Running PostgreSQL migrations..."

# Ensure the schema_migrations table exists to track applied migrations
PGPASSWORD="$POSTGRES_PASSWORD" psql \
    -h "$POSTGRES_HOST" \
    -U "$POSTGRES_USER" \
    -d "$POSTGRES_DB" \
    -c "
CREATE TABLE IF NOT EXISTS schema_migrations (
    version TEXT PRIMARY KEY,
    applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
"

# Get sorted list of migration files (ensures numerical order like 001, 002, etc.)
MIGRATION_FILES=($(ls "$MIGRATIONS_DIR"/*_up.sql 2>/dev/null | sort))

if [ ${#MIGRATION_FILES[@]} -eq 0 ]; then
    echo "No migration files found in $MIGRATIONS_DIR"
    exit 0
fi

# Run each migration if not already applied
for file in "${MIGRATION_FILES[@]}"; do
    version=$(basename "$file" .sql)
    echo "Checking migration: $version"

    # Check if already applied
    if PGPASSWORD="$POSTGRES_PASSWORD" psql \
        -h "$POSTGRES_HOST" \
        -U "$POSTGRES_USER" \
        -d "$POSTGRES_DB" \
        -t \
        -c "SELECT 1 FROM schema_migrations WHERE version='$version';" | grep -q 1; then
        echo "Skipping $version (already applied)"
    else
        echo "Running migration: $version"
        PGPASSWORD="$POSTGRES_PASSWORD" psql \
            -h "$POSTGRES_HOST" \
            -U "$POSTGRES_USER" \
            -d "$POSTGRES_DB" \
            -f "$file"
        
        # Mark as applied
        PGPASSWORD="$POSTGRES_PASSWORD" psql \
            -h "$POSTGRES_HOST" \
            -U "$POSTGRES_USER" \
            -d "$POSTGRES_DB" \
            -c "INSERT INTO schema_migrations (version) VALUES ('$version');"
        echo "Applied $version"
    fi
done

echo "All migrations completed."