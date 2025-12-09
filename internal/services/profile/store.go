package profile

import (
	"database/sql"
)

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

// func (s *Store) GetProfile() ([]types.Profile, error) {
// 	rows, err := s.db.Query("SELECT * FROM profiles WHERE userId = ?", )
// }
