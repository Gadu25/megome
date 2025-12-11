package profile

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

func (s *Store) GetProfile(userId int) ([]types.Profile, error) {
	_, err := s.db.Query("SELECT * FROM profiles WHERE userId = ?", userId)
	if err != nil {
		return nil, err
	}

	// p := new(types.Profile)
	// for rows.Next() {
	// 	p, err = scanRowIntoProfile(rows)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// }

	return nil, nil
}
