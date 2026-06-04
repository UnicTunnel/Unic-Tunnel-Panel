package store

import (
	"context"
	"errors"
	"testing"
	"time"
)

func openTest(t *testing.T) *Store {
	t.Helper()
	s, err := Open(":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s
}

func TestAdminRoundTrip(t *testing.T) {
	s := openTest(t)
	ctx := context.Background()
	if err := s.EnsureAdmin(ctx, "admin", "hash-v1"); err != nil {
		t.Fatalf("ensure: %v", err)
	}
	got, err := s.GetAdmin(ctx, "admin")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Username != "admin" || got.PasswordHash != "hash-v1" {
		t.Fatalf("unexpected: %+v", got)
	}
}

func TestEnsureAdminUpdatesHash(t *testing.T) {
	s := openTest(t)
	ctx := context.Background()
	_ = s.EnsureAdmin(ctx, "admin", "old")
	_ = s.EnsureAdmin(ctx, "admin", "new")
	a, _ := s.GetAdmin(ctx, "admin")
	if a.PasswordHash != "new" {
		t.Fatalf("expected updated hash, got %q", a.PasswordHash)
	}
}

func TestGetAdminNotFound(t *testing.T) {
	s := openTest(t)
	_, err := s.GetAdmin(context.Background(), "ghost")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestSessionLifecycle(t *testing.T) {
	s := openTest(t)
	ctx := context.Background()
	_ = s.EnsureAdmin(ctx, "admin", "h")
	a, _ := s.GetAdmin(ctx, "admin")

	sess, err := s.CreateSession(ctx, a.ID, time.Hour)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if len(sess.ID) != 64 {
		t.Fatalf("expected 64-char hex token, got %d", len(sess.ID))
	}

	got, err := s.GetSession(ctx, sess.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.AdminID != a.ID {
		t.Fatalf("admin id mismatch: %d vs %d", got.AdminID, a.ID)
	}

	if err := s.DeleteSession(ctx, sess.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, err := s.GetSession(ctx, sess.ID); !errors.Is(err, ErrNoSession) {
		t.Fatalf("expected ErrNoSession after delete, got %v", err)
	}
}

func TestSessionExpiryAutoDeletes(t *testing.T) {
	s := openTest(t)
	ctx := context.Background()
	_ = s.EnsureAdmin(ctx, "admin", "h")
	a, _ := s.GetAdmin(ctx, "admin")

	sess, _ := s.CreateSession(ctx, a.ID, -time.Minute) // already expired
	if _, err := s.GetSession(ctx, sess.ID); !errors.Is(err, ErrNoSession) {
		t.Fatalf("expected ErrNoSession for expired, got %v", err)
	}
	// And it should have been deleted from the row store.
	if err := s.DeleteSession(ctx, sess.ID); err != nil {
		t.Fatalf("delete after expiry should be a no-op: %v", err)
	}
}

func TestTunnelUserLifecycle(t *testing.T) {
	s := openTest(t)
	ctx := context.Background()

	u, err := s.CreateTunnelUser(ctx, "Sara — laptop", "u_a3f8c2d1", "secret24chars------------xx")
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if u.Status != "active" {
		t.Fatalf("expected active, got %q", u.Status)
	}

	all, err := s.ListTunnelUsers(ctx)
	if err != nil || len(all) != 1 {
		t.Fatalf("list: err=%v, n=%d", err, len(all))
	}

	got, err := s.GetTunnelUser(ctx, u.ID)
	if err != nil || got.Username != "u_a3f8c2d1" {
		t.Fatalf("get: %+v err=%v", got, err)
	}

	if err := s.RevokeTunnelUser(ctx, u.ID); err != nil {
		t.Fatalf("revoke: %v", err)
	}
	got, _ = s.GetTunnelUser(ctx, u.ID)
	if got.Status != "revoked" || got.RevokedAt == nil {
		t.Fatalf("expected revoked status with timestamp, got %+v", got)
	}

	// Second revoke is a no-op (returns ErrNotFound because WHERE status='active' matches 0 rows).
	if err := s.RevokeTunnelUser(ctx, u.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound on second revoke, got %v", err)
	}
}

func TestGetTunnelUserNotFound(t *testing.T) {
	s := openTest(t)
	_, err := s.GetTunnelUser(context.Background(), 999)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}
