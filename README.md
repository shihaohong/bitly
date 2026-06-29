# bitly

A URL shortener REST API built in Go.
Users register, log in, and get back a JWT they use to create, list, and delete short links.
Anyone can follow a short link without authentication.

## Stack

- **Go** with [Gin](https://github.com/gin-gonic/gin)
- **PostgreSQL** (pgx/v5)
- **JWT** for auth (HS256, 24 h expiry)
- **Docker Compose** for local development

## Endpoints

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | `/auth/register` | - | Create account |
| POST | `/auth/login` | - | Get JWT token |
| GET | `/:code` | - | Redirect to original URL |
| POST | `/api/links` | JWT | Shorten a URL |
| GET | `/api/links` | JWT | List your links |
| DELETE | `/api/links/:code` | JWT | Delete a link |

## Running locally

**With Docker Compose (recommended):**

```bash
cp .env.example .env
docker compose up
```

**Without Docker** (requires a running PostgreSQL instance):

```bash
cp .env.example .env
# Edit .env with your DATABASE_URL, then apply the schema:
psql "$DATABASE_URL" -f migrations/001_initial.sql
go run ./cmd/server/main.go
```

The server starts on `http://localhost:8080`.

## Running tests

Tests require a PostgreSQL instance.
By default they connect to `postgres://bitly:bitly@localhost:5432/bitly_test` - create that database once:

```bash
psql -c "CREATE USER bitly WITH PASSWORD 'bitly';"
psql -c "CREATE DATABASE bitly_test OWNER bitly;"
```

Then run all tests:

```bash
go test -p 1 ./...
```

The `-p 1` flag is required because tests across packages share the same database and must not run concurrently.
To target a different test database, set `TEST_DATABASE_URL` before running.

## Environment variables

| Variable | Description |
|----------|-------------|
| `DATABASE_URL` | PostgreSQL connection string |
| `JWT_SECRET` | Secret key for signing JWTs - change in production |
| `PORT` | HTTP port (default `8080`) |

## Schema

Migrations live in `migrations/` and are applied automatically by Docker Compose on first run.
Short codes are 7-character random alphanumeric strings.
Click counts are incremented atomically on each redirect.
