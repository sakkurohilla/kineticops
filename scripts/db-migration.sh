#!/usr/bin/env bash
set -e

echo "Running PostgreSQL migrations..."

# Example using psql. Adjust DB variables as needed
psql -h "${POSTGRES_HOST:-localhost}" -U "${POSTGRES_USER:-kinetic}" -d "${POSTGRES_DB:-kineticdb}" -f migrations/postgres/001_init.up.sql

echo "Migration completed."
