package personalaccesstokens

import (
	"crypto/sha256"
	"database/sql"
	"errors"
	"fmt"
	"megome/internal/services/types"
	"megome/internal/services/utils"
)

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) GetPATByToken(token string) (types.PATMinified, error) {
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

	var pat types.PATMinified

	err := row.Scan(
		&pat.ID,
		&pat.UserID,
		&pat.Name,
		&pat.TokenHash,
		&pat.RevokedAt,
	)

	if err != nil {
		return types.PATMinified{}, err
	}

	return pat, nil
}

func (s *Store) GetPATs(userId int) ([]types.PersonalAccessToken, error) {
	rows, err := s.db.Query(`
		SELECT id, name, lastUsedAt, revokedAt, createdAt, updatedAt
		FROM personal_access_tokens
		WHERE userId = ?
	`, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pats []types.PersonalAccessToken

	for rows.Next() {
		pat, err := scanPATRows(rows)
		if err != nil {
			return nil, err
		}

		pats = append(pats, pat)
	}

	return pats, rows.Err()
}

func (s *Store) CreatePAT(userId int, name string) (string, error) {

	const maxAttempts = 3

	for i := 0; i < maxAttempts; i++ {

		token, err := utils.GenerateRandomToken("pat_")
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

		if utils.IsMysqlDuplicateKeyError(err) {
			continue
		}

		return "", err
	}

	return "", errors.New("failed to generate unique token")
}

func (s *Store) RevokePAT(userId int, tokenId int) error {

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

func (s *Store) DeletePAT(userId int, tokenId int) error {

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

func scanPATRows(rows *sql.Rows) (types.PersonalAccessToken, error) {
	var pat types.PersonalAccessToken

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
