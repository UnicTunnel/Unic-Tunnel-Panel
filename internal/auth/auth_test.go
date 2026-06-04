package auth

import (
	"errors"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestHashThenVerify(t *testing.T) {
	hash, err := HashPassword("hunter2")
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	if hash == "hunter2" {
		t.Fatal("hash equals plaintext — is bcrypt actually running?")
	}
	if err := VerifyPassword(hash, "hunter2"); err != nil {
		t.Fatalf("verify good: %v", err)
	}
	if err := VerifyPassword(hash, "wrong"); !errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		t.Fatalf("verify bad: expected bcrypt mismatch, got %v", err)
	}
}

func TestHashesAreSalted(t *testing.T) {
	// Same input → different hashes because bcrypt salts internally.
	a, _ := HashPassword("same-input")
	b, _ := HashPassword("same-input")
	if a == b {
		t.Fatal("expected bcrypt hashes to be salted (different across calls)")
	}
	if err := VerifyPassword(a, "same-input"); err != nil {
		t.Fatalf("verify a: %v", err)
	}
	if err := VerifyPassword(b, "same-input"); err != nil {
		t.Fatalf("verify b: %v", err)
	}
}
