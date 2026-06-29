# AGENTS.md

Context for AI agents working in this repo.

## What this is

A URL shortener with account-based link management.
Users register/login to manage their short links; redirect is public.
Backend: Go + Gin. Frontend: React 18 + Vite (embedded into the Go binary).

## Layout

```
cmd/server/main.go               - entrypoint, router wiring
internal/
  auth/                          - register + login, bcrypt + JWT
  links/                         - create, list, delete, redirect + click count
  middleware/auth.go             - JWT validation, injects user_id into gin context
  db/db.go                       - pgxpool setup from DATABASE_URL
  web/                           - Go embed wrapper; dist/ populated by React build
migrations/001_initial.sql       - users + links tables
web/                             - React + Vite source
  src/
    api.js                       - fetch wrapper with JWT injection
    App.jsx                      - React Router setup
    pages/                       - Login.jsx, Register.jsx, Dashboard.jsx
    pages/*.test.jsx             - Vitest + React Testing Library unit tests
```

## Build order

The Go binary embeds the React frontend.
Always build the frontend before building/running Go:

```bash
cd web && npm run build   # outputs to internal/web/dist/
cd .. && go run ./cmd/server
```

For frontend-only work: use `npm run dev` in `web/` (hot reload, proxies `/api` and `/auth` to Go on :8080).

## Deploying changes

```bash
# 1. Build image (multi-stage: Node build → Go embed)
gcloud builds submit --config=cloudbuild.yaml --project=shihao-bitly --substitutions=COMMIT_SHA=latest .

# 2. Deploy
gcloud run deploy bitly --image=us-central1-docker.pkg.dev/shihao-bitly/bitly/server:latest --region=us-central1 --project=shihao-bitly
```

## Running tests

```bash
# Frontend unit tests (19 tests)
cd web && npm test

# Backend (no tests yet)
go test ./...
```

## Key constraints

- `/:code` in Gin is a wildcard route for short link redirects.
  `/login`, `/register`, `/dashboard` are registered as explicit static routes before the wildcard and take priority.
  Any new page route must be added to **both** `internal/web/handler.go` (Go) and `web/src/App.jsx` (React Router).
- Short codes are 7-char random base62 (`crypto/rand`).
  The strings `login`, `register`, `dashboard`, `api`, `auth`, `health` are shadowed by explicit routes - do not allow users to claim these as custom codes.
- Click count is incremented in the same `UPDATE ... RETURNING` that resolves a redirect (atomic).
- JWT claims: `sub` = user UUID, `exp` = 24h. Secret is in Secret Manager under `jwt-secret`.
- Delete is user-scoped: `WHERE short_code = $1 AND user_id = $2`.

## Conventions

- Handler files own HTTP concerns only (binding, status codes, JSON responses).
- Service files own business logic and DB queries.
- No ORM - raw pgx queries.
- Errors wrapped with `fmt.Errorf("context: %w", err)`.

## Local dev

```bash
cp .env.example .env
docker compose up db
psql postgres://bitly:bitly@localhost:5432/bitly -f migrations/001_initial.sql
go run ./cmd/server   # API on :8080

# In a second terminal:
cd web && npm run dev  # UI on :5173
```

## Migrations

Never run automatically. Apply manually via the Cloud SQL Auth Proxy:

```bash
/tmp/cloud-sql-proxy shihao-bitly:us-central1:bitly-db --port=5433 &
DB_PASS=$(gcloud secrets versions access latest --secret=db-password --project=shihao-bitly)
PGPASSWORD="$DB_PASS" psql "host=127.0.0.1 port=5433 dbname=bitly user=bitly" \
  -f migrations/002_your_migration.sql
```

Never modify `migrations/001_initial.sql`. Add new numbered files instead.

## GCP resources

| Resource | Name |
|---|---|
| Project | `shihao-bitly` |
| Region | `us-central1` |
| Cloud Run service | `bitly` |
| Cloud SQL instance | `bitly-db` |
| Artifact Registry | `us-central1-docker.pkg.dev/shihao-bitly/bitly/server` |
| Service account | `bitly-server@shihao-bitly.iam.gserviceaccount.com` |
| Secrets | `jwt-secret`, `db-password` |
| Live URL | https://bitly-910354525392.us-central1.run.app |
