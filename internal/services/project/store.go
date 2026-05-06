package project

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
		`SELECT id, title, description, link, githubLink, status, createdAt, updatedAt
		 FROM projects
		 WHERE userId = ?`,
		userId,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	projects := []types.Project{}

	for rows.Next() {
		var p types.Project

		err := rows.Scan(
			&p.ID,
			&p.Title,
			&p.Description,
			&p.Link,
			&p.GithubLink,
			&p.Status,
			&p.CreatedAt,
			&p.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		projects = append(projects, p)
	}

	if err := rows.Err(); err != nil {
		return nil, err
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

func (s *Store) GetProjectImages(projectIds []int) (map[int][]types.ProjectImage, error) {
	if len(projectIds) == 0 {
		return nil, nil
	}

	// build placeholders (?, ?, ?)
	placeholders := ""
	args := make([]interface{}, len(projectIds))

	for i, id := range projectIds {
		if i > 0 {
			placeholders += ","
		}
		placeholders += "?"
		args[i] = id
	}

	query := `
		SELECT id, projectId, url, type, position, createdAt, updatedAt
		FROM project_images
		WHERE projectId IN (` + placeholders + `)
	`

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[int][]types.ProjectImage)

	for rows.Next() {
		var img types.ProjectImage

		err := rows.Scan(
			&img.ID,
			&img.ProjectID,
			&img.URL,
			&img.Type,
			&img.Position,
			&img.CreatedAt,
			&img.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		result[img.ProjectID] = append(result[img.ProjectID], img)
	}

	return result, rows.Err()
}

func (s *Store) GetProjectTechs(projectIds []int) (map[int][]types.Technology, error) {
	if len(projectIds) == 0 {
		return nil, nil
	}

	// build IN (?, ?, ?)
	placeholders := ""
	args := make([]interface{}, len(projectIds))

	for i, id := range projectIds {
		if i > 0 {
			placeholders += ","
		}
		placeholders += "?"
		args[i] = id
	}

	query := `
		SELECT 
			pt.projectId,
			t.id,
			t.createdByUserId,
			t.name,
			t.slug,
			t.category,
			t.isVerified,
			t.createdAt,
			t.updatedAt
		FROM project_techs pt
		INNER JOIN technologies t ON pt.techId = t.id
		WHERE pt.projectId IN (` + placeholders + `)
	`

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[int][]types.Technology)

	for rows.Next() {
		var (
			projectID int
			tech      types.Technology
		)

		err := rows.Scan(
			&projectID,
			&tech.ID,
			&tech.CreatedByUserId,
			&tech.Name,
			&tech.Slug,
			&tech.Category,
			&tech.IsVerified,
			&tech.CreatedAt,
			&tech.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		result[projectID] = append(result[projectID], tech)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func (s *Store) GetProjectsFull(userId int) ([]types.ProjectFull, error) {
	projects, err := s.GetProjects(userId)
	if err != nil {
		return nil, err
	}

	projectIds := make([]int, 0, len(projects))
	for _, p := range projects {
		projectIds = append(projectIds, p.ID)
	}

	imagesMap, err := s.GetProjectImages(projectIds)
	if err != nil {
		return nil, err
	}

	techsMap, err := s.GetProjectTechs(projectIds) // <-- YOU FORGOT THIS
	if err != nil {
		return nil, err
	}

	result := make([]types.ProjectFull, 0, len(projects))

	for _, p := range projects {
		result = append(result, types.ProjectFull{
			Project:      p,
			Images:       mapImages(imagesMap[p.ID]),
			Technologies: techsMap[p.ID],
		})
	}

	return result, nil
}

func mapImages(images []types.ProjectImage) types.ProjectImages {
	var result types.ProjectImages

	for _, img := range images {
		publicURL := utils.GetPublicFile(img.URL)

		switch img.Type {
		case "cover":
			result.Cover = &publicURL

		case "screenshot":
			result.Screenshots = append(result.Screenshots, publicURL)
		}
	}

	return result
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
