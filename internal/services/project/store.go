package project

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"megome/internal/services/storage"
	"megome/internal/services/types"
	"megome/internal/services/utils"
)

type Store struct {
	db      *sql.DB
	storage *storage.R2Client
}

func NewStore(db *sql.DB, storage *storage.R2Client) *Store {
	return &Store{db: db, storage: storage}
}

func (s *Store) GetProjectById(id int) (types.ProjectFull, error) {
	row := s.db.QueryRow(`
		SELECT id, title, description, link, githubLink, status, isDraft, createdAt, updatedAt
		FROM projects
		WHERE id = ?
	`, id)

	project, err := scanProject(row)
	if err != nil {
		return types.ProjectFull{}, err
	}

	imagesMap, err := s.GetProjectImages([]int{id})
	if err != nil {
		return types.ProjectFull{}, err
	}

	techsMap, err := s.GetProjectTechs([]int{id})
	if err != nil {
		return types.ProjectFull{}, err
	}

	return types.ProjectFull{
		Project:      project,
		Images:       mapImages(imagesMap[id]),
		Technologies: techsMap[id],
	}, nil
}

func (s *Store) GetPublicProjects(userId int) ([]types.ProjectFull, error) {
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

	result := make([]types.ProjectFull, 0, len(projects))

	for _, project := range projects {
		result = append(result, types.ProjectFull{
			Project:      project,
			Images:       mapImages(imagesMap[project.ID]),
			Technologies: techsMap[project.ID],
		})
	}

	return result, nil
}

func (s *Store) GetProjects(userId int) ([]types.Project, error) {
	rows, err := s.db.Query(`
		SELECT id, title, description, link, githubLink, status, isDraft, createdAt, updatedAt
		FROM projects
		WHERE userId = ?
	`, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []types.Project

	for rows.Next() {
		project, err := scanProjectRows(rows)
		if err != nil {
			return nil, err
		}

		projects = append(projects, project)
	}

	return projects, rows.Err()
}

func (s *Store) CreateProject(project types.Project) (types.ProjectFull, error) {
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
		return types.ProjectFull{}, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return types.ProjectFull{}, err
	}

	return s.GetProjectById(int(id))
}

func (s *Store) UpdateProject(id int, project types.Project) (types.ProjectFull, error) {
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
		return types.ProjectFull{}, err
	}

	return s.GetProjectById(id)
}

func (s *Store) DeleteProject(id int) (types.ProjectFull, error) {
	project, err := s.GetProjectById(id)
	if err != nil {
		return types.ProjectFull{}, err
	}

	tx, err := s.db.Begin()
	if err != nil {
		return types.ProjectFull{}, err
	}

	// rollback safety
	defer func() {
		_ = tx.Rollback()
	}()

	// 1. delete DB first (source of truth)
	if _, err = tx.Exec(`
		DELETE FROM project_images
		WHERE projectId = ?
	`, id); err != nil {
		return types.ProjectFull{}, err
	}

	if _, err = tx.Exec(`
		DELETE FROM project_techs
		WHERE projectId = ?
	`, id); err != nil {
		return types.ProjectFull{}, err
	}

	if _, err = tx.Exec(`
		DELETE FROM projects
		WHERE id = ?
	`, id); err != nil {
		return types.ProjectFull{}, err
	}

	if err = tx.Commit(); err != nil {
		return types.ProjectFull{}, err
	}

	// 2. delete storage AFTER commit (eventual consistency model)
	ctx := context.Background()

	for _, url := range project.Images.Screenshots {
		key := utils.ExtractR2Key(url)
		err = s.storage.DeleteObject(ctx, key)
		if err != nil {
			return types.ProjectFull{}, err
		}
	}

	if project.Images.Cover != nil && *project.Images.Cover != "" {
		key := utils.ExtractR2Key(*project.Images.Cover)
		err = s.storage.DeleteObject(ctx, key)
		if err != nil {
			return types.ProjectFull{}, err
		}
	}

	return project, nil
}

func (s *Store) GetProjectImages(projectIds []int) (map[int][]types.ProjectImage, error) {
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

	result := make(map[int][]types.ProjectImage)

	for rows.Next() {
		image, err := scanProjectImage(rows)
		if err != nil {
			return nil, err
		}

		result[image.ProjectID] = append(result[image.ProjectID], image)
	}

	return result, rows.Err()
}

func (s *Store) GetProjectTechs(projectIds []int) (map[int][]types.Technology, error) {
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

	result := make(map[int][]types.Technology)

	for rows.Next() {
		projectID, tech, err := scanTechnology(rows)
		if err != nil {
			return nil, err
		}

		result[projectID] = append(result[projectID], tech)
	}

	return result, rows.Err()
}

func (s *Store) GetProjectsFull(userId int) ([]types.ProjectFull, error) {
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

	result := make([]types.ProjectFull, 0, len(projects))

	for _, project := range projects {
		result = append(result, types.ProjectFull{
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
}) (types.Project, error) {
	var project types.Project

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

func scanProjectRows(rows *sql.Rows) (types.Project, error) {
	var project types.Project

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

func scanProjectImage(rows *sql.Rows) (types.ProjectImage, error) {
	var image types.ProjectImage

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

func scanTechnology(rows *sql.Rows) (int, types.Technology, error) {
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

	return projectID, tech, err
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
