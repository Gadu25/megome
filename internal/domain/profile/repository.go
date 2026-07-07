package profile

import (
	"database/sql"
<<<<<<< HEAD:internal/domain/profile/repository.go
	"megome/internal/pkg/httputil"
=======
	"fmt"
	"megome/internal/services/types"
	"megome/internal/services/utils"
>>>>>>> 24a47b6 (fix: issue on public profile api):internal/services/profile/store.go
	"strings"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

<<<<<<< HEAD:internal/domain/profile/repository.go
func (s *Repository) GetProfile(userId int) (*Profile, error) {
=======
func (s *Store) GetPublicProfile(userId int) (*types.Profile, error) {
	row := s.db.QueryRow("SELECT id, userId, tagline, bio, firstname, lastname, title, birthday, phone, website, location, profileImage, createdAt, updatedAt FROM profiles WHERE userId = ? LIMIT 1", userId)
	return scanRowIntoProfile(row)
}

func (s *Store) GetProfile(userId int) (*types.Profile, error) {
>>>>>>> 24a47b6 (fix: issue on public profile api):internal/services/profile/store.go
	row := s.db.QueryRow("SELECT id, userId, tagline, bio, firstname, lastname, title, birthday, phone, website, location, profileImage, createdAt, updatedAt FROM profiles WHERE userId = ? LIMIT 1", userId)
	p, err := scanRowIntoProfile(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (s *Repository) MakeProfile(profile Profile) error {
	existing, err := s.GetProfile(profile.UserID)
	if err != nil {
		return err
	}

	if existing == nil {
		_, err = s.db.Exec("INSERT INTO profiles (userId, tagline, bio, firstName, lastName, title, birthday, phone, website, location, profileImage) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
			profile.UserID,
			profile.Tagline,
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
		return nil
	}

	query := `
		UPDATE profiles 
		SET tagline = ?, bio = ?, firstName = ?, lastName = ?, title = ?, birthday = ?, phone = ?, website = ?, location = ?, updatedAt = CURRENT_TIMESTAMP
	`

	args := []any{
		profile.Tagline,
		profile.Bio,
		profile.FirstName,
		profile.LastName,
		profile.Title,
		profile.Birthday,
		profile.Phone,
		profile.Website,
		profile.Location,
	}

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

	return nil
}

func (s *Repository) UpsertOAuthProfile(profile Profile) error {
	query := `
		INSERT INTO profiles (
			userId,
			tagline,
			bio,
			firstName,
			lastName,
			title,
			birthday,
			phone,
			website,
			location,
			profileImage
		)
		VALUES (?, ?, "", ?, ?, "", ?, "", "", "", ?)
		ON DUPLICATE KEY UPDATE
			firstName = VALUES(firstName),
			lastName = VALUES(lastName),
			profileImage = VALUES(profileImage),
			updatedAt = CURRENT_TIMESTAMP
	`

	_, err := s.db.Exec(
		query,
		profile.UserID,
		profile.Tagline,
		profile.FirstName,
		profile.LastName,
		nil,
		profile.ProfileImage,
	)

	return err
}

func scanRowIntoProfile(scanner interface{ Scan(dest ...interface{}) error }) (Profile, error) {
	var profile Profile
	err := scanner.Scan(
		&profile.ID,
		&profile.UserID,
		&profile.Tagline,
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
		return Profile{}, err
	}

	if profile.ProfileImage != "" {
		if !strings.HasPrefix(profile.ProfileImage, "https://") && !strings.HasPrefix(profile.ProfileImage, "http://") {
			profile.ProfileImage = httputil.GetPublicFile(profile.ProfileImage)
		}
	}

	return profile, nil
}
