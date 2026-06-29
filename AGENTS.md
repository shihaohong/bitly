# AGENTS.md

Context for AI agents working in this repo.

## What this is

A URL shortener API - think a minimal Bitly clone.
Users register/login to manage their short links; redirect is public.

## Layout

```
cmd/server/main.go          - entrypoint, router wiring
internal/
  auth/                     - register + login, bcrypt + JWT
  links/                    - create, list, delete, redirect + click count
  middleware/auth.go         - JWT validation, injects user_id into gin context
  db/db.go                  - pgxpool setup from DATABASE_URL
migrations/001_initial.sql  - users + links tables
```

## Key facts

- Short codes are 7-char random alphanumeric, generated with `crypto/rand`.
- Click count is incremented in the same `UPDATE ... RETURNING` query that resolves a redirect (atomic, no separate select).
- JWT claims: `sub` = user UUID, `exp` = 24 h.
- Auth middleware extracts `sub` and sets it as `middleware.UserIDKey` in the gin context; link handlers read it from there.
- Delete is user-scoped: `WHERE short_code = $1 AND user_id = $2` - users can only delete their own links.

## Local dev

```bash
cp .env.example .env
docker compose up   # starts postgres + server; migrations run automatically
```

No test suite yet - when adding one, use real postgres (testcontainers or a local instance), not mocks.

## Conventions

- Handler files own HTTP concerns only (binding, status codes, JSON responses).
- Service files own business logic and DB queries.
- No ORM - raw pgx queries.
- Errors are wrapped with `fmt.Errorf("context: %w", err)`.
