package education

import (
	"database/sql"
	"megome/internal/pkg/httputil"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (s *Repository) GetEducationById(id int) (Education, error) {
	row := s.db.QueryRow("SELECT id, userId, school, description, degree, fieldOfStudy, startDate, endDate, isPresent, displayOrder, createdAt, updatedAt FROM education WHERE id = ?", id)
	return scanRowIntoEducation(row)
}

func (s *Repository) GetEducations(userID int) ([]Education, error) {
	rows, err := s.db.Query(
		"SELECT id, userId, school, description, degree, fieldOfStudy, startDate, endDate, isPresent, displayOrder, createdAt, updatedAt FROM education WHERE userId = ? ORDER BY displayOrder ASC, id ASC",
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanRowsIntoEducation(rows)
}

func (s *Repository) CreateEducation(education Education) (Education, error) {
	result, err := s.db.Exec("INSERT INTO education (userId, school, description, degree, fieldOfStudy, startDate, endDate, isPresent) VALUES(?, ?, ?, ?, ?, ?, ?, ?)",
		education.UserID,
		education.School,
		education.Description,
		education.Degree,
		education.FieldOfStudy,
		education.StartDate,
		httputil.NilIfEmpty(education.EndDate),
		education.IsPresent,
	)
	if err != nil {
		return Education{}, err
	}

	id, err := result.LastInsertId()

	if err != nil {
		return Education{}, err
	}

	return s.GetEducationById(int(id))
}

func (s *Repository) UpdateEducation(id int, education Education) (Education, error) {
	_, err := s.db.Exec("UPDATE education SET school = ?, description = ?, degree = ?, fieldOfStudy = ?, startDate = ?, endDate = ?, isPresent = ?, updatedAt = CURRENT_TIMESTAMP WHERE id = ?",
		education.School,
		education.Description,
		education.Degree,
		education.FieldOfStudy,
		education.StartDate,
		httputil.NilIfEmpty(education.EndDate),
		education.IsPresent,
		id,
	)
	if err != nil {
		return Education{}, err
	}

	return s.GetEducationById(id)
}

func (s *Repository) DeleteEducation(id int) (Education, error) {
	cert, err := s.GetEducationById(id)

	if err != nil {
		return Education{}, err
	}

	_, err = s.db.Exec("DELETE FROM education WHERE id = ?", id)
	if err != nil {
		return Education{}, err
	}
	return cert, nil
}

func scanRowIntoEducation(scanner interface{ Scan(dest ...interface{}) error }) (Education, error) {
	var education Education
	err := scanner.Scan(
		&education.ID,
		&education.UserID,
		&education.School,
		&education.Description,
		&education.Degree,
		&education.FieldOfStudy,
		&education.StartDate,
		&education.EndDate,
		&education.IsPresent,
		&education.DisplayOrder,
		&education.CreatedAt,
		&education.UpdatedAt,
	)
	if err != nil {
		return Education{}, err
	}
	return education, nil
}

func (s *Repository) ReorderEducations(userID int, items []ReorderItem) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare("UPDATE education SET displayOrder = ?, updatedAt = CURRENT_TIMESTAMP WHERE id = ? AND userId = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, item := range items {
		_, err = stmt.Exec(item.DisplayOrder, item.ID, userID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func scanRowsIntoEducation(rows *sql.Rows) ([]Education, error) {
	var educations []Education
	for rows.Next() {
		educ, err := scanRowIntoEducation(rows)
		if err != nil {
			return nil, err
		}
		educations = append(educations, educ)
	}
	return educations, rows.Err()
}
