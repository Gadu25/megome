package education

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

func (s *Store) GetEducationById(id int) (*types.Education, error) {
	row := s.db.QueryRow("SELECT id, userId, school, degree, fieldOfStudy, startDate, endDate FROM educations WHERE id = ?", id)
	education := new(types.Education)
	err := row.Scan(
		&education.ID,
		&education.UserID,
		&education.School,
		&education.Degree,
		&education.FieldOfStudy,
		&education.StartDate,
		&education.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return education, nil
}

func (s *Store) GetEducations(userID int) ([]types.Education, error) {
	rows, err := s.db.Query(
		"SELECT id, school, degree, fieldOfStudy, startDate, endDate from educations WHERE userId = ?",
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var educations []types.Education

	for rows.Next() {
		educ, err := scanRowIntoEducation(rows)
		if err != nil {
			return nil, err
		}
		educations = append(educations, educ)
	}
	return educations, nil
}

func (s *Store) CreateEducation(education types.Education) error {
	_, err := s.db.Exec("INSERT INTO educations (userId, school, degree, fieldOfStudy, startDate, endDate) VALUES(?, ?, ?, ?, ?, ?)",
		education.UserID,
		education.School,
		education.Degree,
		education.FieldOfStudy,
		education.StartDate,
		education.EndDate,
	)
	if err != nil {
		return err
	}
	return nil
}

func (s *Store) UpdateEducation(id int, education types.Education) error {
	_, err := s.db.Exec("UPDATE educations SET school = ?, degree = ?, fieldOfStudy = ?, startDate = ?, endDate = ?, updatedAt = CURRENT_TIMESTAMP WHERE id = ?",
		education.School,
		education.Degree,
		education.FieldOfStudy,
		education.StartDate,
		education.EndDate,
		id,
	)
	if err != nil {
		return err
	}
	return nil
}

func (s *Store) DeleteEducation(id int) error {
	_, err := s.GetEducationById(id)
	if err != nil {
		return err
	}
	_, err = s.db.Exec("DELETE FROM educations WHERE id = ?", id)
	if err != nil {
		return err
	}
	return nil
}

func scanRowIntoEducation(rows *sql.Rows) (types.Education, error) {
	var education types.Education

	err := rows.Scan(
		&education.ID,
		&education.School,
		&education.Degree,
		&education.FieldOfStudy,
		&education.StartDate,
		&education.EndDate,
	)
	if err != nil {
		return types.Education{}, err
	}

	return education, nil
}
