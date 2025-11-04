# KineticOps

KineticOps is an open-source infrastructure monitoring stack: a lightweight agent that collects system metrics and logs, a Go backend that ingests, persists and broadcasts metrics, and a React frontend for visualizations and live dashboards.

This top-level README summarizes repository layout, quickstart/development instructions, configuration, and useful files for development and deployment.

---

## Project overview

- Agent (Go): collects system metrics, logs, and other telemetry from hosts. See `agent/`.
- Backend (Go): HTTP API, WebSocket hub, workers (retention, downsampling, metric collector), persistence (Postgres), optional logs in MongoDB, Redis for caching, and Redpanda (Kafka) for messaging. See `backend/`.
- Frontend (React + Vite): UI and dashboards. See `frontend/`.
- Dev infra: `docker-compose.yml` brings up Postgres, MongoDB, Redis and Redpanda for local development.
- Migrations: `scripts/db-migration.sh` applies SQL migrations in `backend/migrations/postgres`.

## Repository layout (important files)

- `agent/` — Agent code, README, build & systemd integration details.
- `backend/` — Server code, `cmd/server`, internal services, repositories, messaging, migrations, and config.
- `frontend/` — React + Vite app for dashboards.
- `docker-compose.yml` — Local development services (Postgres, MongoDB, Redis, Redpanda).
- `scripts/db-migration.sh` — Helper script to apply Postgres migrations.

## What it does (data flow)

1. Agents collect metrics and logs from hosts and send events to the backend (HTTP/API output).
2. Backend persists metrics (Postgres), optionally stores logs in MongoDB, caches/coordination via Redis, and publishes events to Redpanda (Kafka API).
3. Backend runs workers for retention, downsampling, and a metric collector; it also broadcasts live events to connected WebSocket clients.
4. Frontend connects to backend HTTP APIs and subscribes to the WebSocket endpoint for live metric updates.

## Quickstart — local development

Prerequisites:

- Docker & Docker Compose
- Go (matching versions used by modules; backend uses Go 1.25.x, agent uses Go 1.21 in their modules)
- Node.js & npm (for frontend)

1) Start development infra (Postgres, MongoDB, Redis, Redpanda):

```bash
docker-compose up -d
```

2) Create or copy the backend environment file.

The repo expects a `backend/.env` to be present for the migration script and for local backend runs. Create `backend/.env` or `backend/.env.local` from the example below (or the variables described in the Configuration section).

3) Run DB migrations:

```bash
./scripts/db-migration.sh
```

4) Start backend server (development):

```bash
cd backend
go run ./cmd/server
# or build: go build -o kineticops-backend ./cmd/server && ./kineticops-backend
```

5) Start frontend during development:

```bash
cd frontend
npm install
npm run dev
```

6) Build and run the agent (on a test host or locally):

```bash
cd agent
make deps    # if provided in Makefile
make build
sudo ./build/kineticops-agent -c /etc/kineticops-agent/config.yaml
```

Agent can also be installed as a systemd service; see `agent/README.md` for details and config examples.

## Configuration (env vars & config files)

- Agent config: by default the agent reads `/etc/kineticops-agent/config.yaml` (see `agent/README.md` for sample config).
- Backend env: create `backend/.env` containing connection strings and secrets. Typical environment variables to set:

```env
# Postgres
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_USER=akash
POSTGRES_PASSWORD=akash
POSTGRES_DB=kineticops

# Redis
REDIS_ADDR=localhost:6379

# MongoDB (optional for logs)
MONGO_URI=mongodb://localhost:27017

# Redpanda/Kafka brokers (comma-separated)
REDPANDA_BROKERS=localhost:9092

# App
APP_PORT=8080
JWT_SECRET=supersecretjwtkey
# Other vars used by backend config (email, OAuth, etc.)
```

Note: `backend/cmd/server/main.go` and `backend/config` will read configuration via Viper; check `backend/config/config.go` for exact keys required by your local setup.

## Useful commands

- Start infra: `docker-compose up -d`
- Apply migrations: `./scripts/db-migration.sh` (requires backend/.env)
- Run backend: `cd backend && go run ./cmd/server`
- Run frontend (dev): `cd frontend && npm install && npm run dev`
- Build agent: `cd agent && make build`

## Notable implementation details

- Backend uses Fiber for HTTP and the `gofiber/websocket` contrib for WebSocket endpoints.
- Persistence: PostgreSQL (metrics, relational domain models) + MongoDB for log storage.
- Messaging: Redpanda (Kafka API) used to publish/consume metric events; backend also keeps a WebSocket fallback hub.
- Worker processes: retention (garbage collect old metrics), downsampling, metric collector, heartbeat monitor.

## Tests & CI

- Agent README references `make test`; run tests in respective modules if test suites exist.
- CI should run `go vet`, `go test ./...` for Go modules and `npm run lint` / `npm test` for frontend (if tests exist). Consider adding a root-level CI pipeline that builds backend, agent and the frontend.

## Recommended next steps / improvements

- Add `backend/.env.example` to the repo with required env vars and sensible defaults.
- Add a `docker-compose.dev.yml` that includes backend and frontend services wired to the local infra for single-command local dev.
- Add a root-level `README` badge and a short troubleshooting section (common startup errors, missing env vars).
- Add a CONTRIBUTING.md and CODE_OF_CONDUCT for open-source contributions.

## Developer & Maintainer contacts

If you need help, check the module READMEs (`agent/README.md`) and the backend `internal` package docs. For issues, open a GitHub issue in the project repository.

---

Thank you for working with KineticOps — if you want, I can also:

- Create `backend/.env.example` with commonly used environment variables.
- Add a `docker-compose.dev.yml` that includes backend and frontend for full local stack testing.
- Add a `Makefile` / root-level convenience scripts for common dev tasks.

Choose one or more and I will add them now.
