package profile

import (
	"database/sql"
	"megome/internal/services/types"
	"megome/internal/services/utils"
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
		_, err = s.db.Exec("INSERT INTO profiles (userId, bio, firstName, lastName, title, birthday, phone, website, location, profileImage) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
			profile.UserID,
			profile.Bio,
			profile.FirstName,
			profile.LastName,
			profile.Title,
			profile.Birthday,
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
		query := `
			UPDATE profiles 
			SET bio = ?, firstName = ?, lastName = ?, title = ?, birthday = ?, phone = ?, website = ?, location = ?, updatedAt = CURRENT_TIMESTAMP
		`

		args := []any{
			profile.Bio,
			profile.FirstName,
			profile.LastName,
			profile.Title,
			profile.Birthday,
			profile.Phone,
			profile.Website,
			profile.Location,
		}

		// only include if provided
		if profile.ProfileImage != "" {
			query += ", profileImage = ?"
			args = append(args, profile.ProfileImage)
		}

		query += " WHERE userId = ?"
		args = append(args, profile.UserID)

		_, err = s.db.Exec(query, args...)
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
		&profile.FirstName,
		&profile.LastName,
		&profile.Title,
		&profile.Birthday,
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

	if profile.ProfileImage != "" {
		profile.ProfileImage = utils.GetPublicFile(profile.ProfileImage)
	}

	return profile, nil
}
