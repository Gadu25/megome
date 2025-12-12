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

func (s *Store) GetProfile(userId int) (*types.Profile, error) {
	row := s.db.QueryRow("SELECT * FROM profiles WHERE userId = ? LIMIT 1", userId)
	return scanRowIntoProfile(row)
}

func scanRowIntoProfile(row *sql.Row) (*types.Profile, error) {
	profile := new(types.Profile)

	err := row.Scan(
		&profile.ID,
		&profile.UserID,
		&profile.Bio,
		&profile.Phone,
		&profile.Website,
		&profile.Location,
		&profile.ProfileImage,
		&profile.CreatedAt,
		&profile.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return profile, nil
}
