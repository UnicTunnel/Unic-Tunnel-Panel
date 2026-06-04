// Package store is the SQLite persistence layer for the panel.
// It owns the schema and raw CRUD for admins, sessions, and tunnel users.
// Crypto (bcrypt, password generation) lives elsewhere — the store accepts
// already-hashed values and stores them as opaque text.
package store

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"time"

	_ "modernc.org/sqlite"
)

const schema = `
PRAGMA journal_mode = WAL;
PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS admins (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    username      TEXT    NOT NULL UNIQUE,
    password_hash TEXT    NOT NULL,
    created_at    INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS sessions (
    id         TEXT    PRIMARY KEY,
    admin_id   INTEGER NOT NULL REFERENCES admins(id) ON DELETE CASCADE,
    created_at INTEGER NOT NULL,
    expires_at INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS tunnel_users (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    name       TEXT    NOT NULL,
    username   TEXT    NOT NULL UNIQUE,
    password   TEXT    NOT NULL,
    status     TEXT    NOT NULL DEFAULT 'active',
    created_at INTEGER NOT NULL,
    revoked_at INTEGER
);
`

type Store struct{ db *sql.DB }

type Admin struct {
	ID           int64
	Username     string
	PasswordHash string
	CreatedAt    time.Time
}

type Session struct {
	ID        string
	AdminID   int64
	CreatedAt time.Time
	ExpiresAt time.Time
}

type TunnelUser struct {
	ID        int64
	Name      string
	Username  string
	Password  string
	Status    string
	CreatedAt time.Time
	RevokedAt *time.Time
}

var (
	ErrNotFound  = errors.New("not found")
	ErrNoSession = errors.New("no session")
)

// Open opens (or creates) the SQLite DB at the given path and ensures the schema.
// Use ":memory:" for tests.
func Open(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path+"?_pragma=busy_timeout(5000)")
	if err != nil {
		return nil, err
	}
	if _, err := db.Exec(schema); err != nil {
		_ = db.Close()
		return nil, err
	}
	return &Store{db: db}, nil
}

func (s *Store) Close() error { return s.db.Close() }

// EnsureAdmin creates the admin row or updates the password hash if the row exists.
// The caller is responsible for producing passwordHash (e.g. bcrypt).
func (s *Store) EnsureAdmin(ctx context.Context, username, passwordHash string) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO admins(username, password_hash, created_at) VALUES(?,?,?)
		 ON CONFLICT(username) DO UPDATE SET password_hash = excluded.password_hash`,
		username, passwordHash, time.Now().Unix())
	return err
}

func (s *Store) GetAdmin(ctx context.Context, username string) (*Admin, error) {
	var a Admin
	var createdAt int64
	err := s.db.QueryRowContext(ctx,
		`SELECT id, username, password_hash, created_at FROM admins WHERE username = ?`,
		username).Scan(&a.ID, &a.Username, &a.PasswordHash, &createdAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	a.CreatedAt = time.Unix(createdAt, 0)
	return &a, nil
}

func (s *Store) CreateSession(ctx context.Context, adminID int64, ttl time.Duration) (*Session, error) {
	id, err := randomToken(32)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	expires := now.Add(ttl)
	if _, err := s.db.ExecContext(ctx,
		`INSERT INTO sessions(id, admin_id, created_at, expires_at) VALUES(?,?,?,?)`,
		id, adminID, now.Unix(), expires.Unix()); err != nil {
		return nil, err
	}
	return &Session{ID: id, AdminID: adminID, CreatedAt: now, ExpiresAt: expires}, nil
}

// GetSession returns the session if it exists and hasn't expired. Expired sessions
// are deleted on access so the table doesn't grow indefinitely.
func (s *Store) GetSession(ctx context.Context, id string) (*Session, error) {
	var sess Session
	var createdAt, expiresAt int64
	err := s.db.QueryRowContext(ctx,
		`SELECT id, admin_id, created_at, expires_at FROM sessions WHERE id = ?`,
		id).Scan(&sess.ID, &sess.AdminID, &createdAt, &expiresAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNoSession
	}
	if err != nil {
		return nil, err
	}
	sess.CreatedAt = time.Unix(createdAt, 0)
	sess.ExpiresAt = time.Unix(expiresAt, 0)
	if time.Now().After(sess.ExpiresAt) {
		_ = s.DeleteSession(ctx, id)
		return nil, ErrNoSession
	}
	return &sess, nil
}

func (s *Store) DeleteSession(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM sessions WHERE id = ?`, id)
	return err
}

func (s *Store) CreateTunnelUser(ctx context.Context, name, username, password string) (*TunnelUser, error) {
	now := time.Now()
	res, err := s.db.ExecContext(ctx,
		`INSERT INTO tunnel_users(name, username, password, status, created_at) VALUES(?,?,?,?,?)`,
		name, username, password, "active", now.Unix())
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()
	return &TunnelUser{ID: id, Name: name, Username: username, Password: password, Status: "active", CreatedAt: now}, nil
}

func (s *Store) ListTunnelUsers(ctx context.Context) ([]TunnelUser, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, name, username, password, status, created_at, revoked_at
		 FROM tunnel_users ORDER BY id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []TunnelUser
	for rows.Next() {
		var u TunnelUser
		var createdAt int64
		var revokedAt sql.NullInt64
		if err := rows.Scan(&u.ID, &u.Name, &u.Username, &u.Password, &u.Status, &createdAt, &revokedAt); err != nil {
			return nil, err
		}
		u.CreatedAt = time.Unix(createdAt, 0)
		if revokedAt.Valid {
			t := time.Unix(revokedAt.Int64, 0)
			u.RevokedAt = &t
		}
		out = append(out, u)
	}
	return out, rows.Err()
}

func (s *Store) GetTunnelUser(ctx context.Context, id int64) (*TunnelUser, error) {
	var u TunnelUser
	var createdAt int64
	var revokedAt sql.NullInt64
	err := s.db.QueryRowContext(ctx,
		`SELECT id, name, username, password, status, created_at, revoked_at
		 FROM tunnel_users WHERE id = ?`,
		id).Scan(&u.ID, &u.Name, &u.Username, &u.Password, &u.Status, &createdAt, &revokedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	u.CreatedAt = time.Unix(createdAt, 0)
	if revokedAt.Valid {
		t := time.Unix(revokedAt.Int64, 0)
		u.RevokedAt = &t
	}
	return &u, nil
}

// RevokeTunnelUser marks the user revoked. Returns ErrNotFound if the user
// doesn't exist or is already revoked (idempotent from the caller's POV via the
// status they read back).
func (s *Store) RevokeTunnelUser(ctx context.Context, id int64) error {
	res, err := s.db.ExecContext(ctx,
		`UPDATE tunnel_users SET status='revoked', revoked_at=? WHERE id=? AND status='active'`,
		time.Now().Unix(), id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func randomToken(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
