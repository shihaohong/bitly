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
| Registry | Artifact Registry |
| Secrets | Secret Manager |

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
    └── src/
        ├── api.js      fetch wrapper with JWT injection
        ├── App.jsx     React Router setup
        └── pages/      Login, Register, Dashboard
```

## Local development

### Prerequisites

- Go 1.25+
- Node 22+
- Docker (for Postgres)

### 1. Start Postgres

```bash
docker compose up db
```

### 2. Run migrations

```bash
psql postgres://bitly:bitly@localhost:5432/bitly -f migrations/001_initial.sql
```

### 3. Start the API server

```bash
cp .env.example .env
go run ./cmd/server
```

### 4. Start the React dev server (separate terminal)

```bash
cd web && npm install && npm run dev
```

Open `http://localhost:5173`.
Vite proxies `/api` and `/auth` to the Go server on port 8080.

## Frontend tests

```bash
cd web
npm test            # run once
npm run test:watch  # watch mode
```

19 unit tests covering Login, Register, and Dashboard components.

## Building for production (embedded)

The Dockerfile builds in two stages: Node builds the React app, then Go embeds `dist/` into the binary.
To replicate locally:

```bash
cd web && npm run build   # outputs to internal/web/dist/
cd .. && go build ./cmd/server
```

## Deployment (GCP)

GCP project: `shihao-bitly` | Region: `us-central1`

**Build and push image:**
```bash
gcloud builds submit \
  --config=cloudbuild.yaml \
  --project=shihao-bitly \
  --substitutions=COMMIT_SHA=latest .
```

**Deploy to Cloud Run:**
```bash
gcloud run deploy bitly \
  --image=us-central1-docker.pkg.dev/shihao-bitly/bitly/server:latest \
  --region=us-central1 \
  --project=shihao-bitly
```

**Live URL:** https://bitly-910354525392.us-central1.run.app

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

## Adding migrations

```bash
/tmp/cloud-sql-proxy shihao-bitly:us-central1:bitly-db --port=5433 &
DB_PASS=$(gcloud secrets versions access latest --secret=db-password --project=shihao-bitly)
PGPASSWORD="$DB_PASS" psql "host=127.0.0.1 port=5433 dbname=bitly user=bitly" \
  -f migrations/002_your_migration.sql
```
