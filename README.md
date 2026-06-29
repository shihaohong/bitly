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

```bash
cp .env.example .env
docker compose up
```

The server starts on `http://localhost:8080`.

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
