#!/usr/bin/env bash
# Cleanup demo / sample / test data for kineticops
# Dry-run by default. Requires --apply and CONFIRM=YES to actually delete data.
# Supports: Postgres (psql), MongoDB (mongosh), Redis (redis-cli)

set -euo pipefail

HERE=$(cd "$(dirname "$0")" && pwd)

usage() {
  cat <<EOF
Usage: $0 [--apply] [--pg-conn "connection-string"]

By default this script runs in dry-run mode showing counts of candidate tables/collections
that commonly hold demo or sample data. To actually delete/truncate data pass --apply and
set CONFIRM=YES in the environment (this protects against accidental runs).

Examples:
  # Dry-run
  $0

  # Apply (requires CONFIRM=YES)
  CONFIRM=YES $0 --apply --pg-conn "postgres://user:pass@localhost:5432/kineticops?sslmode=disable"

Notes:
  - This script attempts safe, coarse-grained cleanup. Inspect output before running with --apply.
  - Postgres: will try to TRUNCATE a set of candidate tables if available.
  - MongoDB: will try to delete documents from commonly named collections.
  - Redis: will show DB size and optionally FLUSHDB if requested and CONFIRM=YES.
EOF
}

DRY_RUN=1
PG_CONN=""
APPLY=0

while [[ $# -gt 0 ]]; do
  case "$1" in
    --apply) APPLY=1; DRY_RUN=0; shift ;;
    --pg-conn) PG_CONN="$2"; shift 2 ;;
    -h|--help) usage; exit 0 ;;
    *) echo "Unknown arg: $1"; usage; exit 2 ;;
  esac
done

if [[ $APPLY -eq 1 && "${CONFIRM:-}" != "YES" ]]; then
  echo "To actually apply destructive changes set CONFIRM=YES in the environment." >&2
  echo "Example: CONFIRM=YES $0 --apply" >&2
  exit 2
fi

echo "KineticOps cleanup script"
if [[ $DRY_RUN -eq 1 ]]; then
  echo "Mode: dry-run (no destructive actions).";
else
  echo "Mode: APPLY (destructive). CONFIRM=${CONFIRM:-}";
fi

echo
echo "=== Postgres ==="
if command -v psql >/dev/null 2>&1; then
  if [[ -n "$PG_CONN" ]]; then
    PSQL="psql '$PG_CONN' -At -c"
  else
    # Rely on PGHOST/PGUSER/PGDATABASE env vars or ~/.pgpass
    PSQL="psql -At -c"
  fi

  PG_TABLES=(hosts agents installation_tokens logs metrics timeseries_metrics alert_rules synthetic checks deployments)

  for t in "${PG_TABLES[@]}"; do
    set +e
    cnt=$($PSQL "SELECT COUNT(*) FROM ${t};" 2>/dev/null || echo "-")
    rc=$?
    set -e
    if [[ $rc -ne 0 ]]; then
      echo "  ${t}: not present or query failed"
    else
      echo "  ${t}: ${cnt} rows"
      if [[ $APPLY -eq 1 ]]; then
        echo "    -> Truncating ${t} (CASCADE)"
        $PSQL "TRUNCATE TABLE ${t} CASCADE;"
      fi
    fi
  done
else
  echo "psql not found - skipping Postgres checks. Set --pg-conn or ensure psql is on PATH.";
fi

echo
echo "=== MongoDB ==="
if command -v mongosh >/dev/null 2>&1; then
  MONGO_DB=${MONGO_DB:-kineticops}
  echo "Using MongoDB DB: ${MONGO_DB}"

  MONGO_CAND=(agents logs events metrics snapshots installation_tokens)

  for col in "${MONGO_CAND[@]}"; do
    set +e
    cnt=$(mongosh --quiet --eval "db.getSiblingDB('${MONGO_DB}').getCollection('${col}').countDocuments()" 2>/dev/null || echo "-")
    rc=$?
    set -e
    if [[ $rc -ne 0 ]]; then
      echo "  ${col}: not present or query failed"
    else
      echo "  ${col}: ${cnt} documents"
      if [[ $APPLY -eq 1 ]]; then
        echo "    -> Deleting documents from ${col}"
        mongosh --quiet --eval "db.getSiblingDB('${MONGO_DB}').getCollection('${col}').deleteMany({})"
      fi
    fi
  done
else
  echo "mongosh not found - skipping MongoDB checks. Install mongosh or set PATH.";
fi

echo
echo "=== Redis ==="
if command -v redis-cli >/dev/null 2>&1; then
  REDIS_URL=${REDIS_URL:-}
  if [[ -n "$REDIS_URL" ]]; then
    REDIS_CLI=(redis-cli -u "$REDIS_URL")
  else
    REDIS_CLI=(redis-cli)
  fi
  set +e
  size=$(${REDIS_CLI[@]} DBSIZE 2>/dev/null || echo "-")
  rc=$?
  set -e
  if [[ $rc -ne 0 ]]; then
    echo "  redis: lookup failed or redis-cli not reachable"
  else
    echo "  redis: ${size} keys"
    if [[ $APPLY -eq 1 ]]; then
      echo "    -> FLUSHDB"
      ${REDIS_CLI[@]} FLUSHDB
    fi
  fi
else
  echo "redis-cli not found - skipping Redis checks.";
fi

echo
echo "=== Git / Repo sample files ==="
echo "The repository contains example/sample/demo strings and example assets. We do not delete code files automatically here."
echo "Suggested next step: review the candidate files listed in docs/cleanup.md and confirm deletions."

echo
echo "Cleanup script finished. If you ran with --apply, verify services and backups immediately." 
