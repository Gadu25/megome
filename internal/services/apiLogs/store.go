package apilogs

import (
	"database/sql"
	"megome/internal/services/types"
)

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) Create(log types.APIUsageLog) error {
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

func (s *Store) GetByTokenID(tokenID int, limit int, offset int) (types.APIUsageLogWithToken, error) {
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
		return types.APIUsageLogWithToken{}, err
	}
	defer rows.Close()

	var result types.APIUsageLogWithToken
	logs := make([]types.APIUsageLog, 0)

	var token types.PersonalAccessToken
	tokenLoaded := false

	for rows.Next() {
		var log types.APIUsageLog

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
			return types.APIUsageLogWithToken{}, err
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

// func (s *Store) GetRequestCountByUserID(userID int) (int, error) {
// 	var count int

// 	err := s.db.QueryRow(`
// 		SELECT COUNT(*)
// 		FROM api_usage_logs
// 		WHERE userId = ?
// 	`, userID).Scan(&count)

// 	if err != nil {
// 		return 0, err
// 	}

// 	return count, nil
// }

// func (s *Store) GetAverageResponseTimeByUserID(userID int) (float64, error) {
// 	var avg sql.NullFloat64

// 	err := s.db.QueryRow(`
// 		SELECT AVG(responseTimeMs)
// 		FROM api_usage_logs
// 		WHERE userId = ?
// 	`, userID).Scan(&avg)

// 	if err != nil {
// 		return 0, err
// 	}

// 	if !avg.Valid {
// 		return 0, nil
// 	}

// 	return avg.Float64, nil
// }

func (s *Store) GetUserUsageStats(userID int) (types.UserAPIUsageStats, error) {
	var stats types.UserAPIUsageStats

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
		return types.UserAPIUsageStats{}, err
	}

	if avg.Valid {
		stats.AverageResponseMs = avg.Float64
	}

	return stats, nil
}
