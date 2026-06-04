// Package accounts provisions locked-down tunnel-only SSH users on a VPS.
// v1 ships the Provisioner interface + a Stub impl (logs only). The Sudo impl
// that shells out to /usr/local/sbin/tunnel-user.sh lands when deploy is in sight.
package accounts

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"log"
)

// Provisioner manages the OS-level SSH accounts the panel mints.
// Implementations must be safe to call from multiple goroutines.
type Provisioner interface {
	Create(ctx context.Context, username, password string) error
	Lock(ctx context.Context, username string) error
	Delete(ctx context.Context, username string) error
}

// NewStub returns a no-op provisioner that logs what it would have done.
// Use this for local Windows dev where there's no sshd to provision against.
func NewStub() Provisioner { return &stub{} }

type stub struct{}

func (s *stub) Create(_ context.Context, username, password string) error {
	log.Printf("stub: would CREATE %s (pw len=%d)", username, len(password))
	return nil
}

func (s *stub) Lock(_ context.Context, username string) error {
	log.Printf("stub: would LOCK %s", username)
	return nil
}

func (s *stub) Delete(_ context.Context, username string) error {
	log.Printf("stub: would DELETE %s", username)
	return nil
}

// GenerateUsername returns a username like "u_a3f8c2d1": prefix + 8 hex chars
// (32 bits of entropy). useradd on the VPS will reject collisions, so we don't
// pre-check; 4B distinct names is enough headroom.
func GenerateUsername() (string, error) {
	b := make([]byte, 4)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return "u_" + hex.EncodeToString(b), nil
}

// GeneratePassword returns 24 URL-safe random characters from 18 bytes of
// CSPRNG entropy (~144 bits).
func GeneratePassword() (string, error) {
	b := make([]byte, 18)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
