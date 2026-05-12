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
