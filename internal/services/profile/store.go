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

func (s *Store) MakeProfile(profile types.Profile) error {
	existing, err := s.GetProfile(profile.UserID)
	if err != nil {
		_, err = s.db.Exec("INSERT INTO profiles (userId, bio, phone, website, location, profileImage) VALUES (?, ?, ?, ?, ?, ?)",
			profile.UserID,
			profile.Bio,
			profile.Phone,
			profile.Website,
			profile.Location,
			profile.ProfileImage,
		)
		if err != nil {
			return err
		}
	}
	if existing != nil {
		_, err = s.db.Exec("UPDATE profiles SET bio = ?, phone = ?, website = ?, location = ?, profileImage = ?, updatedAt = CURRENT_TIMESTAMP WHERE userId = ?",
			profile.Bio,
			profile.Phone,
			profile.Website,
			profile.Location,
			profile.ProfileImage,
			profile.UserID,
		)
		if err != nil {
			return err
		}
	}

	return nil
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
