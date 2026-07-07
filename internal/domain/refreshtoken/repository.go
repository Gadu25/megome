package refreshtoken

import (
	"crypto/sha256"
	"database/sql"
	"errors"
	"fmt"
	"megome/internal/config"
	"megome/internal/pkg/auth"
	"megome/internal/pkg/httputil"
	"time"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (s *Repository) CreateRefreshToken(userId int) (string, error) {
	for i := 0; i < 2; i++ {
		token, err := httputil.GenerateRandomToken("")
		if err != nil {
			return "", err
		}

		hash := sha256.Sum256([]byte(token))
		hashStr := fmt.Sprintf("%x", hash)

		expiresAt := time.Now().Add(14 * 24 * time.Hour)

		_, err = s.db.Exec("INSERT INTO refresh_tokens (userId, tokenHash, expiresAt) VALUES (?, ?, ?)",
			userId,
			hashStr,
			expiresAt,
		)

		if err == nil {
			return token, nil
		}

		if httputil.IsMysqlDuplicateKeyError(err) {
			continue
		}

		return "", err
	}

	return "", errors.New("failed to generate unique refresh token")
}

func (s *Repository) RefreshRotation(token string) (string, string, error) {
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

	var refreshToken RefreshToken
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

func (s *Repository) LogoutUser(token string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	hash := sha256.Sum256([]byte(token))
	hashStr := fmt.Sprintf("%x", hash)

	_, err = tx.Exec("UPDATE refresh_tokens SET revokedAt = NOW(), updatedAt = CURRENT_TIMESTAMP WHERE tokenHash = ?",
		hashStr,
	)
	if err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}
