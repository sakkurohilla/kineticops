# Cleanup and remove demo / sample / mock data

This document explains how to safely identify and remove demo/sample/mock/test data from the repository and the running development databases for KineticOps.

High-level approach
- Inventory candidate files and DB collections/tables (non-destructive preview).
- Create backups (DB dumps, file copies) before applying destructive operations.
- Run the interactive cleanup script in dry-run mode, review the output.
- When ready, run with CONFIRM=YES and `--apply`.

Files / strings found by the repo inventory (candidates to review):

- `scripts/install-promtail.sh` (contains example REGISTER_URL comments)
- `backend/scripts/agent_batch_send.sh` (uses `agent-demo-token` by default)
- `.promtail-sim/` (simulation artifacts from tester runs)
- `docs/promtail-agent-deploy.md` (contains example LOKI_URL values)
- `agent/README.md` (contains example endpoints like `https://primary.example.com`)
- `frontend` placeholder values and example strings (email placeholders, example.com URLs in several components)
- `backend/internal/api/handlers/host_handler.go` contains a comment "seed with snapshots" — check for programmatic seed behavior
- `backend/migrations/postgres/001_init_all_up.sql` is marked "Clean version - NO sample data" (safe)

What the included cleanup script does
- `scripts/cleanup_databases.sh` (dry-run by default) will:
  - For Postgres: query the counts of candidate tables and optionally TRUNCATE them when `--apply` and `CONFIRM=YES` are provided.
  - For MongoDB: count documents in candidate collections and optionally delete them.
  - For Redis: show key counts and optionally FLUSHDB.

Safety notes
- The script is intentionally conservative: it only targets common table/collection names used by the application and requires explicit CONFIRM=YES to apply.
- Always create DB backups (pg_dump, mongodump, redis rdb snapshot) before running the apply mode.
- Review the list of candidate repo files and confirm whether you want to remove or sanitize them. Many of the matches are just example placeholders (e.g., `you@example.com`) that are harmless.

Recommended next steps
1. Run the script in dry-run mode and paste the output here if you want guidance interpreting it:

   ```bash
   ./scripts/cleanup_databases.sh
   ```

2. If the dry-run shows rows/documents you want removed, backup and then run:

   ```bash
   # Backup first (examples)
   pg_dumpall > /tmp/kineticops-pg-backup.sql
   mongodump --db kineticops --out /tmp/kineticops-mongo-backup

   # Then apply cleanup (dangerous):
   CONFIRM=YES ./scripts/cleanup_databases.sh --apply --pg-conn "postgres://user:pass@localhost:5432/kineticops?sslmode=disable"
   ```

3. Review `docs/cleanup.md` and identify code/files to be removed from the repository. I can prepare a patch to remove or sanitize them after you confirm which items to delete.

If you want, I can now:
- Run the dry-run cleanup script here (non-destructive) and paste the output for review. OR
- Prepare a repo patch that removes/sanitizes identified sample/demo files (you must confirm which ones to delete).

If you want me to proceed automatically with dangerous destructive changes, explicitly type: "I confirm run destructive cleanup" and I will proceed after taking backups where possible.
