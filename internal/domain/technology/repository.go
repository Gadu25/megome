package technology

import (
	"database/sql"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (s *Repository) GetTechnologyById(id int) (*Technology, error) {
	row := s.db.QueryRow("SELECT id, name, slug, createdAt, updatedAt FROM technologies WHERE id = ?", id)
	technology := new(Technology)
	err := row.Scan(
		&technology.ID,
		&technology.Name,
		&technology.Slug,
		&technology.CreatedAt,
		&technology.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return technology, nil
}

func (s *Repository) GetTechnologies(limit int, offset int) ([]Technology, error) {
	rows, err := s.db.Query(
		"SELECT id, createdByUserId, name, slug, category, isVerified, createdAt, updatedAt FROM technologies LIMIT ? OFFSET ?",
		limit,
		offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	technologies := make([]Technology, 0)

	for rows.Next() {
		tech, err := scanRowIntoTechnology(rows)
		if err != nil {
			return nil, err
		}
		technologies = append(technologies, tech)
	}
	return technologies, nil
}

func (s *Repository) CreateTechnology(technology Technology) error {
	_, err := s.db.Exec(
		"INSERT INTO technologies (userId, name, slug) VALUES(?, ?, ?)",
		technology.CreatedByUserId,
		technology.Name,
		technology.Slug,
	)
	if err != nil {
		return err
	}
	return nil
}

func (s *Repository) UpdateTechnology(id int, technology Technology) error {
	_, err := s.db.Exec(
		"UPDATE technologies SET name = ?, slug = ?, updatedAt = CURRENT_TIMESTAMP WHERE id = ?",
		technology.Name,
		technology.Slug,
		id,
	)
	if err != nil {
		return err
	}
	return nil
}

func (s *Repository) DeleteTechnology(id int) error {
	_, err := s.GetTechnologyById(id)
	if err != nil {
		return err
	}
	_, err = s.db.Exec("DELETE FROM technologies WHERE id = ?", id)
	return nil
}

func scanRowIntoTechnology(rows *sql.Rows) (Technology, error) {
	var technology Technology

	err := rows.Scan(
		&technology.ID,
		&technology.CreatedByUserId,
		&technology.Name,
		&technology.Slug,
		&technology.Category,
		&technology.IsVerified,
		&technology.CreatedAt,
		&technology.UpdatedAt,
	)
	if err != nil {
		return Technology{}, err
	}

	return technology, nil
}

func (s *Repository) CountAll() (int, error) {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM technologies").Scan(&count)
	return count, err
}
