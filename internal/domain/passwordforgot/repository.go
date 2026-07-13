package passwordforgot

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"megome/internal/pkg/auth"
	"time"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (s *Repository) SavePasswordResetToken(userId int, token string, exp time.Time) error {
	hash := sha256.Sum256([]byte(token))
	tokenHash := hex.EncodeToString(hash[:])

	query := `
		INSERT INTO password_reset_tokens (userId, tokenHash, expiresAt, createdAt)
		VALUES (?, ?, ?, ?)
	`

	_, err := s.db.Exec(query, userId, tokenHash, exp, time.Now())
	return err
}

func (s *Repository) isTokenValid(token string) (int, error) {
	hash := sha256.Sum256([]byte(token))
	tokenHash := hex.EncodeToString(hash[:])

	var userId int
	var expiresAt time.Time
	var usedAt sql.NullTime

	query := `
		SELECT userId, expiresAt, usedAt
		FROM password_reset_tokens
		WHERE tokenHash = ?
		LIMIT 1
	`
	err := s.db.QueryRow(query, tokenHash).Scan(&userId, &expiresAt, &usedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, errors.New("invalid token")
		}
		return 0, err
	}

	if usedAt.Valid {
		return 0, errors.New("token already used")
	}

	if time.Now().After(expiresAt) {
		return 0, errors.New("token expired")
	}

	return userId, nil
}

func (s *Repository) ChangePassword(token string, password string) error {
	userId, err := s.isTokenValid(token)

	if err != nil {
		return err
	}

	hashedPassword, err := auth.HashedPassword(password)

	if err != nil {
		return err
	}

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec(`
		UPDATE users
		SET password = ?
		WHERE id = ?
	`, string(hashedPassword), userId)

	if err != nil {
		return err
	}

	hash := sha256.Sum256([]byte(token))
	tokenHash := hex.EncodeToString(hash[:])

	result, err := tx.Exec(`
		UPDATE password_reset_tokens
		SET usedAt = ?
		where tokenHash = ?
			AND usedAt IS NULL
	`, time.Now(), tokenHash)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected != 1 {
		return errors.New("failed to mark token as used")
	}

	return tx.Commit()
}
