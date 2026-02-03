package technology

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

func (s *Store) GetTechnologyById(id int) (*types.Technology, error) {
	row := s.db.QueryRow("SELECT id, name, slug, createdAt, updatedAt FROM technologies WHERE id = ?", id)
	technology := new(types.Technology)
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

func (s *Store) GetTechnologies(userID int) ([]types.Technology, error) {
	rows, err := s.db.Query(
		"SELECT id, userId, name, slug, createdAt, updatedAt FROM technologies WHERE userId = ?", userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var technologies []types.Technology

	for rows.Next() {
		tech, err := scanRowIntoTechnology(rows)
		if err != nil {
			return nil, err
		}
		technologies = append(technologies, tech)
	}
	return technologies, nil
}

func (s *Store) CreateTechnology(technology types.Technology) error {
	_, err := s.db.Exec(
		"INSERT INTO technologies (userId, name, slug) VALUES(?, ?, ?)",
		technology.UserID,
		technology.Name,
		technology.Slug,
	)
	if err != nil {
		return err
	}
	return nil
}

func (s *Store) UpdateTechnology(id int, technology types.Technology) error {
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

func (s *Store) DeleteTechnology(id int) error {
	_, err := s.GetTechnologyById(id)
	if err != nil {
		return err
	}
	_, err = s.db.Exec("DELETE FROM technologies WHERE id = ?", id)
	return nil
}

func scanRowIntoTechnology(rows *sql.Rows) (types.Technology, error) {
	var technology types.Technology

	err := rows.Scan(
		&technology.ID,
		&technology.UserID,
		&technology.Name,
		&technology.Slug,
		&technology.CreatedAt,
		&technology.UpdatedAt,
	)
	if err != nil {
		return types.Technology{}, err
	}

	return technology, nil
}
