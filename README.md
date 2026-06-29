# Bitly

A URL shortener with account-based link management, built with Go and React, deployed on GCP.

## Stack

| Layer | Technology |
|---|---|
| Backend | Go, Gin |
| Frontend | React 18, Vite, React Router v6 |
| Database | PostgreSQL (Cloud SQL) |
| Auth | JWT (HS256, 24h), bcrypt |
| Hosting | Cloud Run (`us-central1`) |

## Project structure

```
bitly/
├── cmd/server/         Go entrypoint
├── internal/
│   ├── auth/           register + login handlers and service
│   ├── db/             pgxpool connection helper
│   ├── links/          create / resolve / list / delete
│   ├── middleware/     JWT auth middleware
│   └── web/            Go embed wrapper (dist/ populated by React build)
├── migrations/         SQL migration files
└── web/                React + Vite source
```

## Local development

**Prerequisites:** Go 1.25+, Node 22+, Docker (for Postgres)

```bash
# 1. Start Postgres
docker compose up db

# 2. Run migrations
psql postgres://bitly:bitly@localhost:5432/bitly -f migrations/001_initial.sql

# 3. Start the API server
cp .env.example .env
go run ./cmd/server

# 4. Start the React dev server (separate terminal)
cd web && npm install && npm run dev
```

Open `http://localhost:5173`. Vite proxies `/api` and `/auth` to the Go server on port 8080.

## Tests

**Frontend:**
```bash
cd web && npm test
```

**Backend unit tests** (no database needed):
```bash
go test ./internal/links/...
```

**Backend integration + e2e tests** (requires Postgres via `docker compose up db`):
```bash
go test ./internal/... ./e2e/...
```

## Building for production

The Dockerfile builds in two stages: Node builds the React app, then Go embeds `dist/` into the binary.

```bash
cd web && npm run build   # outputs to internal/web/dist/
cd .. && go build ./cmd/server
```

## Deployment (GCP)

GCP project: `shihao-bitly` | Region: `us-central1` | [Live URL](https://bitly-910354525392.us-central1.run.app)

```bash
# Build and push image
gcloud builds submit \
  --config=cloudbuild.yaml \
  --project=shihao-bitly \
  --substitutions=COMMIT_SHA=latest .

# Deploy to Cloud Run
gcloud run deploy bitly \
  --image=us-central1-docker.pkg.dev/shihao-bitly/bitly/server:latest \
  --region=us-central1 \
  --project=shihao-bitly
```

## API

| Method | Path | Auth | Description |
|---|---|---|---|
| POST | `/auth/register` | - | Create account |
| POST | `/auth/login` | - | Get JWT token |
| GET | `/:code` | - | Redirect to original URL |
| POST | `/api/links` | Bearer | Create short link |
| GET | `/api/links` | Bearer | List your links |
| DELETE | `/api/links/:code` | Bearer | Delete a link |
| GET | `/health` | - | Health check |

## Environment variables

| Variable | Description |
|---|---|
| `DATABASE_URL` | PostgreSQL connection string |
| `JWT_SECRET` | HS256 signing secret |
| `PORT` | HTTP port (default `8080`) |
