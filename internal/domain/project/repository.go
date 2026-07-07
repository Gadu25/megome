package project

import (
	"context"
	"database/sql"
	"fmt"
	"megome/internal/domain/technology"
	"megome/internal/pkg/httputil"
	"megome/internal/pkg/storage"
	"strings"
)

type Repository struct {
	db      *sql.DB
	storage *storage.R2Client
}

func NewRepository(db *sql.DB, storage *storage.R2Client) *Repository {
	return &Repository{db: db, storage: storage}
}

func (s *Repository) GetProjectById(id int) (ProjectFull, error) {
	row := s.db.QueryRow(`
		SELECT id, title, description, link, githubLink, status, isDraft, createdAt, updatedAt
		FROM projects
		WHERE id = ?
	`, id)

	project, err := scanProject(row)
	if err != nil {
		return ProjectFull{}, err
	}

	imagesMap, err := s.GetProjectImages([]int{id})
	if err != nil {
		return ProjectFull{}, err
	}

	techsMap, err := s.GetProjectTechs([]int{id})
	if err != nil {
		return ProjectFull{}, err
	}

	return ProjectFull{
		Project:      project,
		Images:       mapImages(imagesMap[id]),
		Technologies: techsMap[id],
	}, nil
}

func (s *Repository) GetPublicProjectById(id int) (ProjectFull, error) {
	row := s.db.QueryRow(`
		SELECT id, title, description, link, githubLink, status, isDraft, createdAt, updatedAt
		FROM projects
		WHERE id = ?
	`, id)

	project, err := scanProject(row)
	if err != nil {
		return ProjectFull{}, err
	}

	imagesMap, err := s.GetProjectImages([]int{id})
	if err != nil {
		return ProjectFull{}, err
	}

	techsMap, err := s.GetProjectTechs([]int{id})
	if err != nil {
		return ProjectFull{}, err
	}

	return ProjectFull{
		Project:      project,
		Images:       mapImages(imagesMap[id]),
		Technologies: techsMap[id],
	}, nil
}

func (s *Repository) GetPublicProjects(userId int) ([]ProjectFull, error) {
	projects, err := s.GetProjects(userId)
	if err != nil {
		return nil, err
	}

	projectIDs := make([]int, 0, len(projects))

	for _, project := range projects {
		projectIDs = append(projectIDs, project.ID)
	}

	imagesMap, err := s.GetProjectImages(projectIDs)
	if err != nil {
		return nil, err
	}

	techsMap, err := s.GetProjectTechs(projectIDs)
	if err != nil {
		return nil, err
	}

	result := make([]ProjectFull, 0, len(projects))

	for _, project := range projects {
		result = append(result, ProjectFull{
			Project:      project,
			Images:       mapImages(imagesMap[project.ID]),
			Technologies: techsMap[project.ID],
		})
	}

	return result, nil
}

func (s *Repository) GetProjects(userId int) ([]Project, error) {
	rows, err := s.db.Query(`
		SELECT id, title, description, link, githubLink, status, isDraft, createdAt, updatedAt
		FROM projects
		WHERE userId = ?
	`, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []Project

	for rows.Next() {
		project, err := scanProjectRows(rows)
		if err != nil {
			return nil, err
		}

		projects = append(projects, project)
	}

	return projects, rows.Err()
}

func (s *Repository) CreateProject(project Project) (ProjectFull, error) {
	result, err := s.db.Exec(`
		INSERT INTO projects (title, status, description, link, githubLink, userId)
		VALUES (?, ?, ?, ?, ?, ?)
	`,
		project.Title,
		project.Status,
		project.Description,
		project.Link,
		project.GithubLink,
		project.UserID,
	)
	if err != nil {
		return ProjectFull{}, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return ProjectFull{}, err
	}

	return s.GetProjectById(int(id))
}

func (s *Repository) UpdateProject(id int, project Project) (ProjectFull, error) {
	_, err := s.db.Exec(`
		UPDATE projects
		SET
			title = ?,
			description = ?,
			status = ?,
			link = ?,
			githubLink = ?,
			isDraft = ?,
			updatedAt = CURRENT_TIMESTAMP
		WHERE id = ?
	`,
		project.Title,
		project.Description,
		project.Status,
		project.Link,
		project.GithubLink,
		project.IsDraft,
		id,
	)
	if err != nil {
		return ProjectFull{}, err
	}

	return s.GetProjectById(id)
}

func (s *Repository) DeleteProject(id int) (ProjectFull, error) {
	project, err := s.GetProjectById(id)
	if err != nil {
		return ProjectFull{}, err
	}

	tx, err := s.db.Begin()
	if err != nil {
		return ProjectFull{}, err
	}

	defer func() {
		_ = tx.Rollback()
	}()

	if _, err = tx.Exec(`
		DELETE FROM project_images
		WHERE projectId = ?
	`, id); err != nil {
		return ProjectFull{}, err
	}

	if _, err = tx.Exec(`
		DELETE FROM project_techs
		WHERE projectId = ?
	`, id); err != nil {
		return ProjectFull{}, err
	}

	if _, err = tx.Exec(`
		DELETE FROM projects
		WHERE id = ?
	`, id); err != nil {
		return ProjectFull{}, err
	}

	if err = tx.Commit(); err != nil {
		return ProjectFull{}, err
	}

	ctx := context.Background()

	for _, url := range project.Images.Screenshots {
		key := httputil.ExtractR2Key(url)
		err = s.storage.DeleteObject(ctx, key)
		if err != nil {
			return ProjectFull{}, err
		}
	}

	if project.Images.Cover != nil && *project.Images.Cover != "" {
		key := httputil.ExtractR2Key(*project.Images.Cover)
		err = s.storage.DeleteObject(ctx, key)
		if err != nil {
			return ProjectFull{}, err
		}
	}

	return project, nil
}

func (s *Repository) GetProjectImages(projectIds []int) (map[int][]ProjectImage, error) {
	query, args := buildInQuery(projectIds)
	if query == "" {
		return nil, nil
	}

	rows, err := s.db.Query(fmt.Sprintf(`
		SELECT id, projectId, url, type, position, createdAt, updatedAt
		FROM project_images
		WHERE projectId IN (%s)
	`, query), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[int][]ProjectImage)

	for rows.Next() {
		image, err := scanProjectImage(rows)
		if err != nil {
			return nil, err
		}

		result[image.ProjectID] = append(result[image.ProjectID], image)
	}

	return result, rows.Err()
}

func (s *Repository) GetProjectTechs(projectIds []int) (map[int][]technology.Technology, error) {
	query, args := buildInQuery(projectIds)
	if query == "" {
		return nil, nil
	}

	rows, err := s.db.Query(fmt.Sprintf(`
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
		WHERE pt.projectId IN (%s)
	`, query), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[int][]technology.Technology)

	for rows.Next() {
		projectID, tech, err := scanTechnology(rows)
		if err != nil {
			return nil, err
		}

		result[projectID] = append(result[projectID], tech)
	}

	return result, rows.Err()
}

func (s *Repository) GetProjectsFull(userId int) ([]ProjectFull, error) {
	projects, err := s.GetProjects(userId)
	if err != nil {
		return nil, err
	}

	projectIDs := make([]int, 0, len(projects))

	for _, project := range projects {
		projectIDs = append(projectIDs, project.ID)
	}

	imagesMap, err := s.GetProjectImages(projectIDs)
	if err != nil {
		return nil, err
	}

	techsMap, err := s.GetProjectTechs(projectIDs)
	if err != nil {
		return nil, err
	}

	result := make([]ProjectFull, 0, len(projects))

	for _, project := range projects {
		result = append(result, ProjectFull{
			Project:      project,
			Images:       mapImages(imagesMap[project.ID]),
			Technologies: techsMap[project.ID],
		})
	}

	return result, nil
}

func buildInQuery(ids []int) (string, []interface{}) {
	if len(ids) == 0 {
		return "", nil
	}

	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids))

	for i, id := range ids {
		placeholders[i] = "?"
		args[i] = id
	}

	return strings.Join(placeholders, ","), args
}

func scanProject(scanner interface {
	Scan(dest ...interface{}) error
}) (Project, error) {
	var project Project

	err := scanner.Scan(
		&project.ID,
		&project.Title,
		&project.Description,
		&project.Link,
		&project.GithubLink,
		&project.Status,
		&project.IsDraft,
		&project.CreatedAt,
		&project.UpdatedAt,
	)

	return project, err
}

func scanProjectRows(rows *sql.Rows) (Project, error) {
	var project Project

	err := rows.Scan(
		&project.ID,
		&project.Title,
		&project.Description,
		&project.Link,
		&project.GithubLink,
		&project.Status,
		&project.IsDraft,
		&project.CreatedAt,
		&project.UpdatedAt,
	)

	return project, err
}

func scanProjectImage(rows *sql.Rows) (ProjectImage, error) {
	var image ProjectImage

	err := rows.Scan(
		&image.ID,
		&image.ProjectID,
		&image.URL,
		&image.Type,
		&image.Position,
		&image.CreatedAt,
		&image.UpdatedAt,
	)

	return image, err
}

func scanTechnology(rows *sql.Rows) (int, technology.Technology, error) {
	var (
		projectID int
		tech      technology.Technology
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

	return projectID, tech, err
}

func mapImages(images []ProjectImage) ProjectImages {
	var result ProjectImages

	for _, img := range images {
		publicURL := httputil.GetPublicFile(img.URL)

		switch img.Type {
		case "cover":
			result.Cover = &publicURL

		case "screenshot":
			result.Screenshots = append(result.Screenshots, publicURL)
		}
	}

	return result
}

func (s *Repository) GetProjectImageByID(id int) (ProjectImage, error) {
	row := s.db.QueryRow(`
		SELECT id, projectId, url, type, position, createdAt, updatedAt
		FROM project_images
		WHERE id = ?
	`, id)

	var img ProjectImage
	err := row.Scan(
		&img.ID,
		&img.ProjectID,
		&img.URL,
		&img.Type,
		&img.Position,
		&img.CreatedAt,
		&img.UpdatedAt,
	)
	if err != nil {
		return ProjectImage{}, err
	}

	return img, nil
}

func (s *Repository) GetProjectImagesByProjectID(projectId int) ([]ProjectImage, error) {
	rows, err := s.db.Query(`
		SELECT id, projectId, url, type, position, createdAt, updatedAt
		FROM project_images
		WHERE projectId = ?
		ORDER BY 
			CASE type
				WHEN 'cover' THEN 0
				WHEN 'demo' THEN 1
				WHEN 'screenshot' THEN 2
			END,
			position ASC
	`, projectId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	images := []ProjectImage{}

	for rows.Next() {
		var img ProjectImage
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
		images = append(images, img)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return images, nil
}

func (s *Repository) AddProjectImage(img ProjectImage) (ProjectImage, error) {
	result, err := s.db.Exec(`
		INSERT INTO project_images (projectId, url, type, position)
		VALUES (?, ?, ?, ?)
	`,
		img.ProjectID,
		img.URL,
		img.Type,
		img.Position,
	)
	if err != nil {
		return ProjectImage{}, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return ProjectImage{}, err
	}

	return s.GetProjectImageByID(int(id))
}

func (s *Repository) DeleteProjectImage(id int) error {
	_, err := s.db.Exec("DELETE FROM project_images WHERE id = ?", id)
	return err
}

func (s *Repository) SetProjectCover(projectId int, img ProjectImage) (ProjectImage, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return ProjectImage{}, err
	}
	defer tx.Rollback()

	_, err = tx.Exec(`
		DELETE FROM project_images
		WHERE projectId = ? AND type = 'cover'
	`, projectId)
	if err != nil {
		return ProjectImage{}, err
	}

	result, err := tx.Exec(`
		INSERT INTO project_images (projectId, url, type)
		VALUES (?, ?, 'cover')
	`, projectId, img.URL)
	if err != nil {
		return ProjectImage{}, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return ProjectImage{}, err
	}

	if err := tx.Commit(); err != nil {
		return ProjectImage{}, err
	}

	return s.GetProjectImageByID(int(id))
}

func (s *Repository) UpdateProjectImagePositions(projectId int, ids []int) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for index, id := range ids {
		_, err := tx.Exec(`
			UPDATE project_images
			SET position = ?
			WHERE id = ? AND projectId = ?
		`, index, id, projectId)

		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *Repository) GetProjectTechById(id int) (*ProjectTech, error) {
	row := s.db.QueryRow("SELECT id, projectId, techId FROM project_techs WHERE id = ?", id)
	projectTech := new(ProjectTech)
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

func (s *Repository) CreateProjectTech(projectTech ProjectTech) error {
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

func (s *Repository) CreateProjectTechBatch(projectID int, techIDs []int) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	stmt, err := tx.Prepare(`INSERT INTO project_techs (projectId, techId) VALUES (?, ?)`)
	if err != nil {
		tx.Rollback()
		return err
	}

	defer stmt.Close()

	for _, techID := range techIDs {
		_, err = stmt.Exec(projectID, techID)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (s *Repository) DelteProjectTech(id int) error {
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
