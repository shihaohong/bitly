package e2e

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shihaohong/bitly/internal/auth"
	"github.com/shihaohong/bitly/internal/links"
	"github.com/shihaohong/bitly/internal/middleware"
	"github.com/shihaohong/bitly/internal/testhelper"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func newRouter(pool *pgxpool.Pool) http.Handler {
	authSvc := auth.NewService(pool)
	authH := auth.NewHandler(authSvc)

	linksSvc := links.NewService(pool)
	linksH := links.NewHandler(linksSvc)

	r := gin.New()
	r.POST("/auth/register", authH.Register)
	r.POST("/auth/login", authH.Login)
	r.GET("/:code", linksH.Redirect)

	api := r.Group("/api", middleware.Auth())
	api.POST("/links", linksH.Create)
	api.GET("/links", linksH.List)
	api.DELETE("/links/:code", linksH.Delete)

	return r
}

func request(t *testing.T, r http.Handler, method, path string, body interface{}, token string) *httptest.ResponseRecorder {
	t.Helper()

	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatalf("encode body: %v", err)
		}
	}

	req := httptest.NewRequest(method, path, &buf)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func decodeJSON(t *testing.T, w *httptest.ResponseRecorder, v interface{}) {
	t.Helper()
	if err := json.NewDecoder(w.Body).Decode(v); err != nil {
		t.Fatalf("decode JSON: %v\nbody: %s", err, w.Body.String())
	}
}

// loginAs registers a new account then logs in, returning the JWT.
func loginAs(t *testing.T, r http.Handler, email, password string) string {
	t.Helper()

	request(t, r, http.MethodPost, "/auth/register",
		map[string]string{"email": email, "password": password}, "")

	w := request(t, r, http.MethodPost, "/auth/login",
		map[string]string{"email": email, "password": password}, "")

	if w.Code != http.StatusOK {
		t.Fatalf("login failed (status %d): %s", w.Code, w.Body.String())
	}

	var resp struct {
		Token string `json:"token"`
	}
	decodeJSON(t, w, &resp)
	return resp.Token
}

// --- Auth tests ---

func TestAuth_Register(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")
	pool := testhelper.NewPool(t)
	r := newRouter(pool)

	t.Run("success", func(t *testing.T) {
		w := request(t, r, http.MethodPost, "/auth/register",
			map[string]string{"email": "alice@example.com", "password": "password123"}, "")
		if w.Code != http.StatusCreated {
			t.Errorf("status=%d, want 201\nbody: %s", w.Code, w.Body.String())
		}
	})

	t.Run("duplicate email", func(t *testing.T) {
		request(t, r, http.MethodPost, "/auth/register",
			map[string]string{"email": "bob@example.com", "password": "password123"}, "")
		w := request(t, r, http.MethodPost, "/auth/register",
			map[string]string{"email": "bob@example.com", "password": "password123"}, "")
		if w.Code != http.StatusConflict {
			t.Errorf("status=%d, want 409", w.Code)
		}
	})

	t.Run("invalid email", func(t *testing.T) {
		w := request(t, r, http.MethodPost, "/auth/register",
			map[string]string{"email": "notanemail", "password": "password123"}, "")
		if w.Code != http.StatusBadRequest {
			t.Errorf("status=%d, want 400", w.Code)
		}
	})

	t.Run("password too short", func(t *testing.T) {
		w := request(t, r, http.MethodPost, "/auth/register",
			map[string]string{"email": "short@example.com", "password": "abc"}, "")
		if w.Code != http.StatusBadRequest {
			t.Errorf("status=%d, want 400", w.Code)
		}
	})
}

func TestAuth_Login(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")
	pool := testhelper.NewPool(t)
	r := newRouter(pool)

	request(t, r, http.MethodPost, "/auth/register",
		map[string]string{"email": "alice@example.com", "password": "password123"}, "")

	t.Run("success returns token", func(t *testing.T) {
		w := request(t, r, http.MethodPost, "/auth/login",
			map[string]string{"email": "alice@example.com", "password": "password123"}, "")
		if w.Code != http.StatusOK {
			t.Errorf("status=%d, want 200\nbody: %s", w.Code, w.Body.String())
		}
		var resp map[string]string
		decodeJSON(t, w, &resp)
		if resp["token"] == "" {
			t.Fatal("expected non-empty token")
		}
	})

	t.Run("wrong password", func(t *testing.T) {
		w := request(t, r, http.MethodPost, "/auth/login",
			map[string]string{"email": "alice@example.com", "password": "wrong"}, "")
		if w.Code != http.StatusUnauthorized {
			t.Errorf("status=%d, want 401", w.Code)
		}
	})

	t.Run("unknown email", func(t *testing.T) {
		w := request(t, r, http.MethodPost, "/auth/login",
			map[string]string{"email": "ghost@example.com", "password": "password123"}, "")
		if w.Code != http.StatusUnauthorized {
			t.Errorf("status=%d, want 401", w.Code)
		}
	})
}

// --- Links tests ---

func TestLinks_Create(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")
	pool := testhelper.NewPool(t)
	r := newRouter(pool)
	token := loginAs(t, r, "alice@example.com", "password123")

	t.Run("success", func(t *testing.T) {
		w := request(t, r, http.MethodPost, "/api/links",
			map[string]string{"url": "https://example.com"}, token)
		if w.Code != http.StatusCreated {
			t.Errorf("status=%d, want 201\nbody: %s", w.Code, w.Body.String())
		}
		var link struct {
			ShortCode  string `json:"short_code"`
			ClickCount int    `json:"click_count"`
		}
		decodeJSON(t, w, &link)
		if len(link.ShortCode) != 7 {
			t.Errorf("short_code %q: len=%d, want 7", link.ShortCode, len(link.ShortCode))
		}
		if link.ClickCount != 0 {
			t.Errorf("click_count=%d, want 0", link.ClickCount)
		}
	})

	t.Run("no auth", func(t *testing.T) {
		w := request(t, r, http.MethodPost, "/api/links",
			map[string]string{"url": "https://example.com"}, "")
		if w.Code != http.StatusUnauthorized {
			t.Errorf("status=%d, want 401", w.Code)
		}
	})

	t.Run("invalid url", func(t *testing.T) {
		w := request(t, r, http.MethodPost, "/api/links",
			map[string]string{"url": "not-a-url"}, token)
		if w.Code != http.StatusBadRequest {
			t.Errorf("status=%d, want 400", w.Code)
		}
	})

	t.Run("missing url field", func(t *testing.T) {
		w := request(t, r, http.MethodPost, "/api/links", map[string]string{}, token)
		if w.Code != http.StatusBadRequest {
			t.Errorf("status=%d, want 400", w.Code)
		}
	})
}

func TestLinks_FullFlow(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")
	pool := testhelper.NewPool(t)
	r := newRouter(pool)
	token := loginAs(t, r, "alice@example.com", "password123")

	// Create a link.
	w := request(t, r, http.MethodPost, "/api/links",
		map[string]string{"url": "https://example.com"}, token)
	if w.Code != http.StatusCreated {
		t.Fatalf("create: status=%d\nbody: %s", w.Code, w.Body.String())
	}
	var link struct {
		ShortCode string `json:"short_code"`
	}
	decodeJSON(t, w, &link)

	// List shows the link with click_count=0.
	w = request(t, r, http.MethodGet, "/api/links", nil, token)
	if w.Code != http.StatusOK {
		t.Fatalf("list: status=%d", w.Code)
	}
	var listed []struct {
		ShortCode  string `json:"short_code"`
		ClickCount int64  `json:"click_count"`
	}
	decodeJSON(t, w, &listed)
	if len(listed) != 1 {
		t.Fatalf("expected 1 link, got %d", len(listed))
	}
	if listed[0].ClickCount != 0 {
		t.Errorf("click_count=%d before redirect, want 0", listed[0].ClickCount)
	}

	// Redirect returns 301 and increments click count.
	w = request(t, r, http.MethodGet, "/"+link.ShortCode, nil, "")
	if w.Code != http.StatusMovedPermanently {
		t.Errorf("redirect: status=%d, want 301", w.Code)
	}
	if loc := w.Header().Get("Location"); loc != "https://example.com" {
		t.Errorf("Location=%q, want %q", loc, "https://example.com")
	}

	// List now shows click_count=1.
	w = request(t, r, http.MethodGet, "/api/links", nil, token)
	decodeJSON(t, w, &listed)
	if listed[0].ClickCount != 1 {
		t.Errorf("click_count=%d after redirect, want 1", listed[0].ClickCount)
	}

	// Delete the link.
	w = request(t, r, http.MethodDelete, "/api/links/"+link.ShortCode, nil, token)
	if w.Code != http.StatusNoContent {
		t.Errorf("delete: status=%d, want 204\nbody: %s", w.Code, w.Body.String())
	}

	// List is now empty.
	w = request(t, r, http.MethodGet, "/api/links", nil, token)
	decodeJSON(t, w, &listed)
	if len(listed) != 0 {
		t.Errorf("expected 0 links after delete, got %d", len(listed))
	}

	// Redirect on deleted link returns 404.
	w = request(t, r, http.MethodGet, "/"+link.ShortCode, nil, "")
	if w.Code != http.StatusNotFound {
		t.Errorf("redirect after delete: status=%d, want 404", w.Code)
	}
}

func TestLinks_UserIsolation(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")
	pool := testhelper.NewPool(t)
	r := newRouter(pool)

	tokenA := loginAs(t, r, "alice@example.com", "password123")
	tokenB := loginAs(t, r, "bob@example.com", "password123")

	// Alice creates a link.
	w := request(t, r, http.MethodPost, "/api/links",
		map[string]string{"url": "https://alice.com"}, tokenA)
	var aliceLink struct {
		ShortCode string `json:"short_code"`
	}
	decodeJSON(t, w, &aliceLink)

	// Bob cannot see Alice's link.
	w = request(t, r, http.MethodGet, "/api/links", nil, tokenB)
	var bobList []interface{}
	decodeJSON(t, w, &bobList)
	if len(bobList) != 0 {
		t.Errorf("bob sees %d links, want 0", len(bobList))
	}

	// Bob cannot delete Alice's link.
	w = request(t, r, http.MethodDelete, "/api/links/"+aliceLink.ShortCode, nil, tokenB)
	if w.Code != http.StatusNotFound {
		t.Errorf("bob deleting alice's link: status=%d, want 404", w.Code)
	}

	// Alice's link still exists.
	w = request(t, r, http.MethodGet, "/api/links", nil, tokenA)
	var aliceList []interface{}
	decodeJSON(t, w, &aliceList)
	if len(aliceList) != 1 {
		t.Errorf("alice sees %d links after bob's failed delete, want 1", len(aliceList))
	}
}

func TestLinks_InvalidToken(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")
	pool := testhelper.NewPool(t)
	r := newRouter(pool)

	w := request(t, r, http.MethodGet, "/api/links", nil, "not.a.valid.token")
	if w.Code != http.StatusUnauthorized {
		t.Errorf("status=%d, want 401", w.Code)
	}
}
