package links_test

import (
	"context"
	"testing"

	"github.com/shihaohong/bitly/internal/links"
	"github.com/shihaohong/bitly/internal/testhelper"
)

func TestService_Create(t *testing.T) {
	pool := testhelper.NewPool(t)
	userID := testhelper.CreateUser(t, pool, "alice@example.com")
	svc := links.NewService(pool)
	ctx := context.Background()

	link, err := svc.Create(ctx, userID, "https://example.com")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if len(link.ShortCode) != 7 {
		t.Errorf("short_code %q: len=%d, want 7", link.ShortCode, len(link.ShortCode))
	}
	if link.OriginalURL != "https://example.com" {
		t.Errorf("original_url=%q, want %q", link.OriginalURL, "https://example.com")
	}
	if link.ClickCount != 0 {
		t.Errorf("click_count=%d, want 0", link.ClickCount)
	}
}

func TestService_Resolve(t *testing.T) {
	pool := testhelper.NewPool(t)
	userID := testhelper.CreateUser(t, pool, "alice@example.com")
	svc := links.NewService(pool)
	ctx := context.Background()

	link, _ := svc.Create(ctx, userID, "https://example.com")

	url, err := svc.Resolve(ctx, link.ShortCode)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if url != "https://example.com" {
		t.Errorf("url=%q, want %q", url, "https://example.com")
	}

	// Click count should increment on each resolve.
	listed, _ := svc.List(ctx, userID)
	if listed[0].ClickCount != 1 {
		t.Errorf("click_count=%d after 1 resolve, want 1", listed[0].ClickCount)
	}

	_, _ = svc.Resolve(ctx, link.ShortCode)
	listed, _ = svc.List(ctx, userID)
	if listed[0].ClickCount != 2 {
		t.Errorf("click_count=%d after 2 resolves, want 2", listed[0].ClickCount)
	}

	_, err = svc.Resolve(ctx, "nosuchcode")
	if err == nil {
		t.Fatal("expected error for unknown code, got nil")
	}
}

func TestService_List(t *testing.T) {
	pool := testhelper.NewPool(t)
	alice := testhelper.CreateUser(t, pool, "alice@example.com")
	bob := testhelper.CreateUser(t, pool, "bob@example.com")
	svc := links.NewService(pool)
	ctx := context.Background()

	got, err := svc.List(ctx, alice)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected 0 links for new user, got %d", len(got))
	}

	_, _ = svc.Create(ctx, alice, "https://a.com")
	_, _ = svc.Create(ctx, alice, "https://b.com")
	_, _ = svc.Create(ctx, bob, "https://c.com")

	got, _ = svc.List(ctx, alice)
	if len(got) != 2 {
		t.Errorf("expected 2 links for alice, got %d", len(got))
	}

	got, _ = svc.List(ctx, bob)
	if len(got) != 1 {
		t.Errorf("expected 1 link for bob, got %d", len(got))
	}
}

func TestService_Delete(t *testing.T) {
	pool := testhelper.NewPool(t)
	alice := testhelper.CreateUser(t, pool, "alice@example.com")
	bob := testhelper.CreateUser(t, pool, "bob@example.com")
	svc := links.NewService(pool)
	ctx := context.Background()

	link, _ := svc.Create(ctx, alice, "https://example.com")

	if err := svc.Delete(ctx, bob, link.ShortCode); err == nil {
		t.Fatal("expected error when another user deletes the link, got nil")
	}

	if err := svc.Delete(ctx, alice, link.ShortCode); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	if err := svc.Delete(ctx, alice, link.ShortCode); err == nil {
		t.Fatal("expected error deleting already-deleted link, got nil")
	}
}
