package passwordForgot

import (
	"database/sql"
	"crypto/sha256"
	"encoding/hex"
	"time"
)

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) SavePasswordResetToken(userId int, token string, exp time.Time) error {
	hash := sha256.Sum256([]byte(token))
	tokenHash := hex.EncodeToString(hash[:])

	query := `
		INSERT INTO password_reset_tokens (userId, tokenHash, expiresAt, createdAt)
		VALUES (?, ?, ?, ?)
	`

	_, err := s.db.Exec(query, userId, tokenHash, exp, time.Now())
	return err
}
