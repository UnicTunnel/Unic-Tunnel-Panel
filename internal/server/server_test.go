package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/UnicTunnel/Unic-Tunnel-Panel/internal/accounts"
	"github.com/UnicTunnel/Unic-Tunnel-Panel/internal/auth"
	"github.com/UnicTunnel/Unic-Tunnel-Panel/internal/store"
)

func newTestServer(t *testing.T) *Server {
	t.Helper()
	st, err := store.Open(":memory:")
	if err != nil {
		t.Fatalf("store: %v", err)
	}
	t.Cleanup(func() { _ = st.Close() })

	hash, _ := auth.HashPassword("hunter2")
	if err := st.EnsureAdmin(context.Background(), "admin", hash); err != nil {
		t.Fatalf("ensure: %v", err)
	}
	return New(Deps{
		Store:       st,
		Provisioner: accounts.NewStub(),
		SSHHost:     "1.2.3.4",
		SSHPort:     2222,
	})
}

func TestLoginPageRenders(t *testing.T) {
	srv := newTestServer(t)
	r := httptest.NewRequest("GET", "/login", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("status: %d", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, `name="password"`) {
		t.Fatalf("login form missing password field; body=%q", body)
	}
}

func TestDashboardRequiresAuth(t *testing.T) {
	srv := newTestServer(t)
	r := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, r)
	if w.Code != http.StatusSeeOther {
		t.Fatalf("status: expected 303 redirect, got %d", w.Code)
	}
	if got := w.Header().Get("Location"); got != "/login" {
		t.Fatalf("location: %q", got)
	}
}

func TestLoginFlowAndCreateUser(t *testing.T) {
	srv := newTestServer(t)

	// Step 1: log in.
	form := strings.NewReader("username=admin&password=hunter2")
	r := httptest.NewRequest("POST", "/login", form)
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, r)
	if w.Code != http.StatusSeeOther {
		t.Fatalf("login status: expected 303, got %d (body=%s)", w.Code, w.Body.String())
	}
	cookies := w.Result().Cookies()
	if len(cookies) == 0 || cookies[0].Name != "unic_session" {
		t.Fatalf("expected unic_session cookie, got %v", cookies)
	}
	sessCookie := cookies[0]

	// Step 2: dashboard renders with session cookie.
	r2 := httptest.NewRequest("GET", "/", nil)
	r2.AddCookie(sessCookie)
	w2 := httptest.NewRecorder()
	srv.ServeHTTP(w2, r2)
	if w2.Code != http.StatusOK {
		t.Fatalf("dashboard status: %d", w2.Code)
	}
	if !strings.Contains(w2.Body.String(), "Tunnel users") {
		t.Fatalf("dashboard missing 'Tunnel users' heading")
	}

	// Step 3: create a tunnel user.
	form3 := strings.NewReader("name=Sara%20laptop")
	r3 := httptest.NewRequest("POST", "/users", form3)
	r3.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r3.AddCookie(sessCookie)
	w3 := httptest.NewRecorder()
	srv.ServeHTTP(w3, r3)
	if w3.Code != http.StatusSeeOther {
		t.Fatalf("create status: expected 303, got %d (body=%s)", w3.Code, w3.Body.String())
	}
	loc := w3.Header().Get("Location")
	if !strings.HasPrefix(loc, "/users/") {
		t.Fatalf("expected redirect to /users/<id>, got %q", loc)
	}

	// Step 4: user detail page shows a unic:// link.
	r4 := httptest.NewRequest("GET", loc, nil)
	r4.AddCookie(sessCookie)
	w4 := httptest.NewRecorder()
	srv.ServeHTTP(w4, r4)
	if w4.Code != http.StatusOK {
		t.Fatalf("user detail status: %d", w4.Code)
	}
	if !strings.Contains(w4.Body.String(), "unic://") {
		t.Fatalf("user detail page missing unic:// link")
	}
}

func TestLoginWithWrongPassword(t *testing.T) {
	srv := newTestServer(t)
	form := strings.NewReader("username=admin&password=wrong")
	r := httptest.NewRequest("POST", "/login", form)
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 re-rendered login form, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Wrong username or password") {
		t.Fatal("expected error message in re-rendered form")
	}
	if len(w.Result().Cookies()) > 0 {
		t.Fatal("no session cookie should be set on bad login")
	}
}
