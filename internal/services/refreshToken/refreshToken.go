package refreshToken

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"math/rand"
	"megome/config"
	"megome/internal/services/auth"
	"megome/internal/services/types"
	"time"

	"github.com/go-sql-driver/mysql"
)

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

func generateRandomToken() (string, error) {
	b := make([]byte, 32) // 256-bit token
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func (s *Store) CreateRefreshToken(userId int) (string, error) {
	// up to three tries if ever generated hash was not unique (very small chance to happen)
	for i := 0; i < 2; i++ {
		token, err := generateRandomToken()
		if err != nil {
			return "", err
		}

		// Hash it
		hash := sha256.Sum256([]byte(token))
		hashStr := fmt.Sprintf("%x", hash) // convert to hex string for DB storage

		// Set expiration 14 days from now
		expiresAt := time.Now().Add(14 * 24 * time.Hour)

		_, err = s.db.Exec("INSERT INTO refresh_tokens (userId, tokenHash, expiresAt) VALUES (?, ?, ?)",
			userId,
			hashStr,
			expiresAt,
		)

		if err == nil {
			return token, nil
		}

		if isDuplicateKeyError(err) {
			continue
		}

		return "", err
	}

	return "", errors.New("failed to generate unique refresh token")
}

func (s *Store) RefreshRotation(token string) (string, string, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return "", "", err
	}
	defer tx.Rollback()

	hash := sha256.Sum256([]byte(token))
	hashStr := fmt.Sprintf("%x", hash)

	row := tx.QueryRow("SELECT id, userId, tokenHash, expiresAt, revokedAt FROM refresh_tokens WHERE tokenHash = ?",
		hashStr,
	)

	var refreshToken types.RefreshToken
	err = row.Scan(
		&refreshToken.ID,
		&refreshToken.UserId,
		&refreshToken.TokenHash,
		&refreshToken.ExpiresAt,
		&refreshToken.RevokedAt,
	)
	if err != nil {
		return "", "", err
	}

	if refreshToken.RevokedAt.Valid {
		// reuse detected → revoke all sessions
		tx.Exec("UPDATE refresh_tokens SET revokedAt = NOW() WHERE userId = ?", refreshToken.UserId)
		if err := tx.Commit(); err != nil {
			return "", "", err
		}
		return "", "", errors.New("Refresh token is already revoked")
	}

	if time.Now().After(refreshToken.ExpiresAt) {
		return "", "", errors.New("Refresh token is expired")
	}

	_, err = tx.Exec("UPDATE refresh_tokens SET revokedAt = NOW(), updatedAt = CURRENT_TIMESTAMP WHERE id = ?",
		refreshToken.ID,
	)
	if err != nil {
		return "", "", err
	}

	newRefreshToken, err := s.CreateRefreshToken(refreshToken.UserId)
	if err != nil {
		return "", "", err
	}

	if err := tx.Commit(); err != nil {
		return "", "", err
	}

	secret := []byte(config.Envs.JWTSecret)
	newAccessToken, err := auth.CreateJWT(secret, refreshToken.UserId)

	return newRefreshToken, newAccessToken, nil
}

func isDuplicateKeyError(err error) bool {
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) {
		// reserved error number for duplicate entry
		return mysqlErr.Number == 1062
	}
	return false
}
