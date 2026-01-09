package experience

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

func (s *Store) GetExperienceById(id int) (*types.Experience, error) {
	row := s.db.QueryRow("SELECT * FROM experiences WHERE id = ?", id)
	experience := new(types.Experience)
	err := row.Scan(
		&experience.ID,
		&experience.UserID,
		&experience.Title,
		&experience.Company,
		&experience.StartDate,
		&experience.EndDate,
		&experience.Description,
		&experience.CreatedAt,
		&experience.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return experience, nil
}

func (s *Store) GetExperiences(userID int) ([]types.Experience, error) {
	rows, err := s.db.Query(
		"SELECT id, userId, title, company, startDate, endDate, description, createdAt, updatedAt FROM experiences WHERE userId = ?",
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var experiences []types.Experience

	for rows.Next() {
		exp, err := scanRowIntoExperience(rows)
		if err != nil {
			return nil, err
		}
		experiences = append(experiences, exp)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return experiences, nil
}

func (s *Store) CreateExperience(experience types.Experience) error {
	_, err := s.db.Exec("INSERT INTO experiences (userId, title, company, startDate, endDate, description) VALUES (?, ?, ?, ?, ?, ?)",
		experience.UserID,
		experience.Title,
		experience.Company,
		experience.StartDate,
		nilIfEmpty(experience.EndDate),
		experience.Description,
	)
	if err != nil {
		return err
	}
	return nil
}

func (s *Store) UpdateExperience(id int, experience types.Experience) error {
	_, err := s.db.Exec("UPDATE experiences SET title = ?, company = ?, startDate = ?, endDate = ?, description = ?, updatedAt = CURRENT_TIMESTAMP WHERE id = ?",
		experience.Title,
		experience.Company,
		experience.StartDate,
		experience.EndDate,
		experience.Description,
		id,
	)
	if err != nil {
		return err
	}
	return nil
}

func (s *Store) DeleteExperience(id int) error {
	_, err := s.GetExperienceById(id)
	if err != nil {
		return err
	}
	_, err = s.db.Exec("DELETE FROM experiences WHERE id = ?", id)
	if err != nil {
		return err
	}
	return nil
}

func scanRowIntoExperience(rows *sql.Rows) (types.Experience, error) {
	var experience types.Experience

	err := rows.Scan(
		&experience.ID,
		&experience.UserID,
		&experience.Title,
		&experience.Company,
		&experience.StartDate,
		&experience.EndDate,
		&experience.Description,
		&experience.CreatedAt,
		&experience.UpdatedAt,
	)
	if err != nil {
		return types.Experience{}, err
	}

	return experience, nil
}

func nilIfEmpty(s *string) *string {
	if s == nil || *s == "" {
		return nil
	}
	return s
}
