#!/usr/bin/env bash
set -e

# Get the directory of the script
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Source the .env file in backend
source "$SCRIPT_DIR/../backend/.env"

echo "Running PostgreSQL migrations..."

PGPASSWORD="$POSTGRES_PASSWORD" psql \
    -h "$POSTGRES_HOST" \
    -U "$POSTGRES_USER" \
    -d "$POSTGRES_DB" \
    -f "$SCRIPT_DIR/../backend/migrations/postgres/002_users_up.sql"

echo "Migration completed."
