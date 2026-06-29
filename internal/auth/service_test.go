package auth_test

import (
	"context"
	"testing"

	"github.com/shihaohong/bitly/internal/auth"
	"github.com/shihaohong/bitly/internal/testhelper"
)

func TestService_Register(t *testing.T) {
	pool := testhelper.NewPool(t)
	svc := auth.NewService(pool)
	ctx := context.Background()

	if err := svc.Register(ctx, "alice@example.com", "password123"); err != nil {
		t.Fatalf("Register: %v", err)
	}

	if err := svc.Register(ctx, "alice@example.com", "other"); err == nil {
		t.Fatal("expected error for duplicate email, got nil")
	}
}

func TestService_Login(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")

	pool := testhelper.NewPool(t)
	svc := auth.NewService(pool)
	ctx := context.Background()

	_ = svc.Register(ctx, "alice@example.com", "password123")

	tests := []struct {
		name    string
		email   string
		pass    string
		wantErr bool
	}{
		{"valid credentials", "alice@example.com", "password123", false},
		{"wrong password", "alice@example.com", "badpass", true},
		{"unknown email", "nobody@example.com", "password123", true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			token, err := svc.Login(ctx, tc.email, tc.pass)
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if token == "" {
				t.Fatal("expected non-empty token")
			}
		})
	}
}
