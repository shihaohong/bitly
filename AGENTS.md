# AGENTS.md

Context for AI agents working in this repo.

## What this is

A URL shortener API - think a minimal Bitly clone.
Users register/login to manage their short links; redirect is public.

## Layout

```
cmd/server/main.go               - entrypoint, router wiring
internal/
  auth/                          - register + login, bcrypt + JWT
  links/                         - create, list, delete, redirect + click count
  middleware/auth.go              - JWT validation, injects user_id into gin context
  db/db.go                       - pgxpool setup from DATABASE_URL
  testhelper/db.go               - shared test DB setup (NewPool, CreateUser)
migrations/001_initial.sql       - users + links tables
internal/links/generate_test.go  - unit tests (no DB)
internal/auth/service_test.go    - auth service integration tests
internal/links/service_test.go   - links service integration tests
e2e/api_test.go                  - full HTTP stack tests via httptest
.github/workflows/test.yml       - CI: postgres service + go test -p 1 -race ./...
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

Run tests with:

```bash
go test -p 1 -race ./...
```

`-p 1` is required - tests across packages share the `bitly_test` postgres database and must not run concurrently.
The test database defaults to `postgres://bitly:bitly@localhost:5432/bitly_test`; override with `TEST_DATABASE_URL`.

**Test layout:**
- `internal/links/generate_test.go` (`package links`) - unit test for `generateCode`, no DB needed.
- `internal/auth/service_test.go`, `internal/links/service_test.go` - service-level integration tests against a real DB.
- `e2e/api_test.go` - full HTTP stack tests using `httptest.NewRecorder` and a wired-up gin router.
- `internal/testhelper/db.go` - `NewPool(t)` applies the schema idempotently and truncates tables before/after each test; `CreateUser(t, pool, email)` inserts a user directly for tests that only need a valid `user_id`.

Always use real postgres for tests - no mocks.

## Conventions

- Handler files own HTTP concerns only (binding, status codes, JSON responses).
- Service files own business logic and DB queries.
- No ORM - raw pgx queries.
- Errors are wrapped with `fmt.Errorf("context: %w", err)`.
