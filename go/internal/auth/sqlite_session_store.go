package auth

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"time"
)

var _ SessionStore = (*SQLiteSessionStore)(nil)

// SQLiteSessionStore persists sessions in a SQL database.
// Tokens are stored as HMAC-SHA256 digests so the raw refresh token never hits disk.
type SQLiteSessionStore struct {
	db  *sql.DB
	key []byte
}

// NewSQLiteSessionStore creates a session store that HMACs tokens before storage.
// The key should be the application's JWT secret (or derived material).
func NewSQLiteSessionStore(db *sql.DB, key []byte) *SQLiteSessionStore {
	return &SQLiteSessionStore{db: db, key: key}
}

func (s *SQLiteSessionStore) hashToken(token string) string {
	mac := hmac.New(sha256.New, s.key)
	mac.Write([]byte(token))
	return hex.EncodeToString(mac.Sum(nil))
}

func (s *SQLiteSessionStore) Create(ctx context.Context, session *Session) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO session (id, user_id, token, expires_at, created_at) VALUES (?, ?, ?, ?, ?)`,
		session.ID, session.UserID, s.hashToken(session.Token), session.ExpiresAt, session.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("inserting session: %w", err)
	}
	return nil
}

func (s *SQLiteSessionStore) FindByToken(ctx context.Context, token string) (*Session, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT id, user_id, expires_at, created_at FROM session WHERE token = ?`,
		s.hashToken(token),
	)

	var sess Session
	if err := row.Scan(&sess.ID, &sess.UserID, &sess.ExpiresAt, &sess.CreatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("session not found")
		}
		return nil, fmt.Errorf("scanning session: %w", err)
	}
	sess.Token = token
	return &sess, nil
}

func (s *SQLiteSessionStore) Revoke(ctx context.Context, token string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM session WHERE token = ?`, s.hashToken(token))
	if err != nil {
		return fmt.Errorf("revoking session: %w", err)
	}
	return nil
}

func (s *SQLiteSessionStore) RevokeAllForUser(ctx context.Context, userID string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM session WHERE user_id = ?`, userID)
	if err != nil {
		return fmt.Errorf("revoking all sessions for user: %w", err)
	}
	return nil
}

func (s *SQLiteSessionStore) DeleteExpired(ctx context.Context) (int64, error) {
	result, err := s.db.ExecContext(ctx, `DELETE FROM session WHERE expires_at < ?`, time.Now())
	if err != nil {
		return 0, fmt.Errorf("deleting expired sessions: %w", err)
	}
	n, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("getting rows affected: %w", err)
	}
	return n, nil
}
