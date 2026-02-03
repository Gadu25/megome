package projecttech

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

func (s *Store) GetProjectTechById(id int) (*types.ProjectTech, error) {
	row := s.db.QueryRow("SELECT id, projectId, techId FROM project_techs WHERE id = ?", id)
	projectTech := new(types.ProjectTech)
	err := row.Scan(
		&projectTech.ID,
		&projectTech.ProjectID,
		&projectTech.TechID,
	)
	if err != nil {
		return nil, err
	}
	return projectTech, nil
}

func (s *Store) CreateProjectTech(projectTech types.ProjectTech) error {
	_, err := s.db.Exec(
		"INSERT into project_techs (projectId, techId) VALUES(?, ?)",
		projectTech.ProjectID,
		projectTech.TechID,
	)
	if err != nil {
		return err
	}
	return nil
}

func (s *Store) DelteProjectTech(id int) error {
	_, err := s.GetProjectTechById(id)
	if err != nil {
		return err
	}
	_, err = s.db.Exec(
		"DELETE FROM project_techs WHERE id = ?",
		id,
	)
	if err != nil {
		return err
	}
	return nil
}
