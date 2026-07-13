package apilog

import (
	"database/sql"
	"megome/internal/domain/personalaccesstoken"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (s *Repository) Create(log APIUsageLog) error {
	_, err := s.db.Exec(`
		INSERT INTO api_usage_logs (
			userId,
			tokenId,
			endpoint,
			method,
			statusCode,
			ipAddress,
			userAgent,
			responseTimeMs
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`,
		log.UserID,
		log.TokenID,
		log.Endpoint,
		log.Method,
		log.StatusCode,
		log.IPAddress,
		log.UserAgent,
		log.ResponseTimeMs,
	)

	return err
}

func (s *Repository) GetByTokenID(tokenID int, limit int, offset int) (APIUsageLogWithToken, error) {
	rows, err := s.db.Query(`
		SELECT
			l.id,
			l.userId,
			l.tokenId,
			l.endpoint,
			l.method,
			l.statusCode,
			l.ipAddress,
			l.userAgent,
			l.responseTimeMs,
			l.createdAt,

			t.id,
			t.userId,
			t.name,
			t.tokenHash,
			t.lastUsedAt,
			t.revokedAt,
			t.createdAt,
			t.updatedAt

		FROM api_usage_logs l
		JOIN personal_access_tokens t ON t.id = l.tokenId
		WHERE l.tokenId = ?
		ORDER BY l.createdAt DESC
		LIMIT ? OFFSET ?
	`, tokenID, limit, offset)

	if err != nil {
		return APIUsageLogWithToken{}, err
	}
	defer rows.Close()

	var result APIUsageLogWithToken
	logs := make([]APIUsageLog, 0)

	var token personalaccesstoken.PersonalAccessToken
	tokenLoaded := false

	for rows.Next() {
		var log APIUsageLog

		err := rows.Scan(
			&log.ID,
			&log.UserID,
			&log.TokenID,
			&log.Endpoint,
			&log.Method,
			&log.StatusCode,
			&log.IPAddress,
			&log.UserAgent,
			&log.ResponseTimeMs,
			&log.CreatedAt,

			&token.ID,
			&token.UserID,
			&token.Name,
			&token.TokenHash,
			&token.LastUsedAt,
			&token.RevokedAt,
			&token.CreatedAt,
			&token.UpdatedAt,
		)

		if err != nil {
			return APIUsageLogWithToken{}, err
		}

		logs = append(logs, log)

		if !tokenLoaded {
			tokenLoaded = true
		}
	}

	result.Logs = logs
	result.Token = token

	return result, nil
}

func (s *Repository) GetUserUsageStats(userID int) (UserAPIUsageStats, error) {
	var stats UserAPIUsageStats

	var avg sql.NullFloat64

	err := s.db.QueryRow(`
		SELECT
			COUNT(*),
			AVG(responseTimeMs)
		FROM api_usage_logs
		WHERE userId = ?
	`, userID).Scan(
		&stats.RequestCount,
		&avg,
	)

	if err != nil {
		return UserAPIUsageStats{}, err
	}

	if avg.Valid {
		stats.AverageResponseMs = avg.Float64
	}

	return stats, nil
}
