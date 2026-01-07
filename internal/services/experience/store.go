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

func (s *Store) CreateExperience(experience types.Experience) error {
	_, err := s.db.Exec("INSERT INTO experiences (userId, title, company, startDate, endDate, description) VALUES (?, ?, ?, ?, ?, ?)",
		experience.UserID,
		experience.Title,
		experience.Company,
		experience.StartDate,
		experience.EndDate,
		experience.Description,
	)
	if err != nil {
		return err
	}
	return nil
}

func (s *Store) UpdateExperience(id int, experience types.Experience) error {
	return nil
}
