package personalaccesstoken

import (
	"crypto/sha256"
	"database/sql"
	"errors"
	"fmt"
	"megome/internal/pkg/httputil"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (s *Repository) GetPATByToken(token string) (PATMinified, error) {
	row := s.db.QueryRow(`
		SELECT
			id,
			userId,
			name,
			tokenHash,
			revokedAt
		FROM personal_access_tokens
		WHERE tokenHash = ?
		LIMIT 1
	`, token)

	var pat PATMinified

	err := row.Scan(
		&pat.ID,
		&pat.UserID,
		&pat.Name,
		&pat.TokenHash,
		&pat.RevokedAt,
	)

	if err != nil {
		return PATMinified{}, err
	}

	return pat, nil
}

func (s *Repository) GetPATs(userId int) ([]PersonalAccessToken, error) {
	rows, err := s.db.Query(`
		SELECT id, name, lastUsedAt, revokedAt, createdAt, updatedAt
		FROM personal_access_tokens
		WHERE userId = ?
	`, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pats []PersonalAccessToken

	for rows.Next() {
		pat, err := scanPATRows(rows)
		if err != nil {
			return nil, err
		}

		pats = append(pats, pat)
	}

	return pats, rows.Err()
}

func (s *Repository) CreatePAT(userId int, name string) (string, error) {

	const maxAttempts = 3

	for i := 0; i < maxAttempts; i++ {

		token, err := httputil.GenerateRandomToken("pat_")
		if err != nil {
			return "", err
		}

		hash := sha256.Sum256([]byte(token))
		hashStr := fmt.Sprintf("%x", hash)

		_, err = s.db.Exec(`
			INSERT INTO personal_access_tokens
			(userId, name, tokenHash)
			VALUES (?, ?, ?)
		`,
			userId,
			name,
			hashStr,
		)

		if err == nil {
			return token, nil
		}

		if httputil.IsMysqlDuplicateKeyError(err) {
			continue
		}

		return "", err
	}

	return "", errors.New("failed to generate unique token")
}

func (s *Repository) RevokePAT(userId int, tokenId int) error {

	result, err := s.db.Exec(`
		UPDATE personal_access_tokens
		SET revokedAt = CURRENT_TIMESTAMP
		WHERE id = ?
		AND userId = ?
		AND revokedAt IS NULL
	`,
		tokenId,
		userId,
	)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("token not found or already revoked")
	}

	return nil
}

func (s *Repository) DeletePAT(userId int, tokenId int) error {

	result, err := s.db.Exec(`
		DELETE FROM personal_access_tokens
		WHERE id = ?
		AND userId = ?
		AND revokedAt IS NOT NULL
	`,
		tokenId,
		userId,
	)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("token must be revoked before deletion")
	}

	return nil
}

func scanPATRows(rows *sql.Rows) (PersonalAccessToken, error) {
	var pat PersonalAccessToken

	err := rows.Scan(
		&pat.ID,
		&pat.Name,
		&pat.LastUsedAt,
		&pat.RevokedAt,
		&pat.CreatedAt,
		&pat.UpdatedAt,
	)

	return pat, err
}

func (s *Repository) GetTokenCountByUserID(userID int) (int, error) {
	var count int

	err := s.db.QueryRow(`
		SELECT COUNT(*)
		FROM personal_access_tokens
		WHERE userId = ?
	`, userID).Scan(&count)

	if err != nil {
		return 0, err
	}

	return count, nil
}
