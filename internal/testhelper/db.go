package testhelper

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

const schema = `
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE IF NOT EXISTS users (
	id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	email         TEXT NOT NULL UNIQUE,
	password_hash TEXT NOT NULL,
	created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS links (
	id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	short_code   TEXT NOT NULL UNIQUE,
	original_url TEXT NOT NULL,
	user_id      UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	click_count  BIGINT NOT NULL DEFAULT 0,
	created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_links_short_code ON links(short_code);
CREATE INDEX IF NOT EXISTS idx_links_user_id ON links(user_id);
`

// NewPool connects to the test database, ensures the schema exists, truncates
// all tables to give each test a clean slate, and registers cleanup.
func NewPool(t *testing.T) *pgxpool.Pool {
	t.Helper()

	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://bitly:bitly@localhost:5432/bitly_test"
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("connect to test db: %v", err)
	}

	if _, err := pool.Exec(ctx, schema); err != nil {
		pool.Close()
		t.Fatalf("apply schema: %v", err)
	}

	truncate(t, pool)
	t.Cleanup(func() {
		truncate(t, pool)
		pool.Close()
	})

	return pool
}

// CreateUser inserts a test user with a known password ("testpass") and returns the UUID.
func CreateUser(t *testing.T, pool *pgxpool.Pool, email string) string {
	t.Helper()

	hash, err := bcrypt.GenerateFromPassword([]byte("testpass"), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	var id string
	if err := pool.QueryRow(context.Background(),
		`INSERT INTO users (email, password_hash) VALUES ($1, $2) RETURNING id`,
		email, string(hash),
	).Scan(&id); err != nil {
		t.Fatalf("create test user: %v", err)
	}

	return id
}

func truncate(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	if _, err := pool.Exec(context.Background(), "TRUNCATE links, users CASCADE"); err != nil {
		t.Fatalf("truncate: %v", err)
	}
}
