package project

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

func (s *Store) GetProjectById(id int) (types.Project, error) {
	row := s.db.QueryRow("SELECT id, title, description, link, githubLink, status, createdAt, updatedAt FROM projects WHERE id = ?", id)

	var project types.Project
	err := row.Scan(
		&project.ID,
		&project.Title,
		&project.Description,
		&project.Link,
		&project.GithubLink,
		&project.Status,
		&project.CreatedAt,
		&project.UpdatedAt,
	)
	if err != nil {
		return types.Project{}, err
	}

	return project, nil
}

func (s *Store) GetProjects(userId int) ([]types.Project, error) {
	rows, err := s.db.Query(
		"SELECT id, title, description, link, githubLink, status FROM projects WHERE userId = ?",
		userId,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	projects := make([]types.Project, 0)

	for rows.Next() {
		project, err := scanRowIntoProject(rows)
		if err != nil {
			return nil, err
		}
		projects = append(projects, project)
	}
	return projects, nil
}

func (s *Store) CreateProject(project types.Project) (types.Project, error) {
	result, err := s.db.Exec(
		"INSERT into projects (title, description, link, githubLink, userId) VALUES(?, ?, ?, ?, ?)",
		project.Title,
		project.Description,
		project.Link,
		project.GithubLink,
		project.UserID,
	)
	if err != nil {
		return types.Project{}, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return types.Project{}, err
	}

	return s.GetProjectById(int(id))
}

func (s *Store) UpdateProject(id int, project types.Project) (types.Project, error) {
	_, err := s.db.Exec(
		"UPDATE projects SET title = ?, description = ?, link = ?, githubLink = ?, isDraft = ?, updatedAt = CURRENT_TIMESTAMP WHERE id = ?",
		project.Title,
		project.Description,
		project.Link,
		project.GithubLink,
		project.IsDraft,
		id,
	)
	if err != nil {
		return types.Project{}, err
	}
	return s.GetProjectById(id)
}

func (s *Store) DeleteProject(id int) (types.Project, error) {
	project, err := s.GetProjectById(id)

	if err != nil {
		return types.Project{}, err
	}

	_, err = s.db.Exec("DELETE FROM projects WHERE id = ?", id)
	if err != nil {
		return types.Project{}, err
	}

	return project, nil
}

func scanRowIntoProject(rows *sql.Rows) (types.Project, error) {
	var project types.Project
	err := rows.Scan(
		&project.ID,
		&project.Title,
		&project.Description,
		&project.Link,
		&project.GithubLink,
		&project.Status,
	)
	if err != nil {
		return types.Project{}, err
	}
	return project, nil
}
