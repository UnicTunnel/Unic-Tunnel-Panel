package accounts

import (
	"context"
	"regexp"
	"testing"
)

func TestGenerateUsernameFormat(t *testing.T) {
	re := regexp.MustCompile(`^u_[0-9a-f]{8}$`)
	seen := make(map[string]bool, 50)
	for i := 0; i < 50; i++ {
		u, err := GenerateUsername()
		if err != nil {
			t.Fatalf("gen: %v", err)
		}
		if !re.MatchString(u) {
			t.Fatalf("bad format: %q", u)
		}
		seen[u] = true
	}
	if len(seen) != 50 {
		t.Fatalf("expected 50 unique usernames, got %d (CSPRNG flake?)", len(seen))
	}
}

func TestGeneratePasswordFormat(t *testing.T) {
	re := regexp.MustCompile(`^[A-Za-z0-9_-]{24}$`)
	seen := make(map[string]bool, 50)
	for i := 0; i < 50; i++ {
		p, err := GeneratePassword()
		if err != nil {
			t.Fatalf("gen: %v", err)
		}
		if !re.MatchString(p) {
			t.Fatalf("bad format: %q", p)
		}
		seen[p] = true
	}
	if len(seen) != 50 {
		t.Fatalf("expected 50 unique passwords, got %d (CSPRNG flake?)", len(seen))
	}
}

func TestStubNoOps(t *testing.T) {
	p := NewStub()
	ctx := context.Background()
	if err := p.Create(ctx, "u_test", "secret"); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if err := p.Lock(ctx, "u_test"); err != nil {
		t.Fatalf("Lock: %v", err)
	}
	if err := p.Delete(ctx, "u_test"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
}
