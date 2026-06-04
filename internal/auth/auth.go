// Package auth wraps bcrypt for admin password hashing/verification.
// Lives outside store so the store layer has no crypto dependency.
package auth

import "golang.org/x/crypto/bcrypt"

// HashPassword returns a bcrypt hash suitable for storing in admins.password_hash.
func HashPassword(password string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// VerifyPassword reports whether the password matches the stored hash.
// Returns nil on match, bcrypt.ErrMismatchedHashAndPassword on mismatch.
func VerifyPassword(hash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}
