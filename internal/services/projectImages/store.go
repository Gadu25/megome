package projectimages

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

func (s *Store) GetProjectImageByID(id int) (types.ProjectImage, error) {
	row := s.db.QueryRow(`
		SELECT id, projectId, url, type, position, createdAt, updatedAt
		FROM project_images
		WHERE id = ?
	`, id)

	var img types.ProjectImage
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
		return types.ProjectImage{}, err
	}

	return img, nil
}

func (s *Store) GetProjectImages(projectId int) ([]types.ProjectImage, error) {
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

	images := []types.ProjectImage{}

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
		images = append(images, img)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return images, nil
}

func (s *Store) AddProjectImage(img types.ProjectImage) (types.ProjectImage, error) {
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
		return types.ProjectImage{}, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return types.ProjectImage{}, err
	}

	return s.GetProjectImageByID(int(id))
}

func (s *Store) DeleteProjectImage(id int) error {
	_, err := s.db.Exec("DELETE FROM project_images WHERE id = ?", id)
	return err
}

func (s *Store) SetProjectCover(projectId int, img types.ProjectImage) (types.ProjectImage, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return types.ProjectImage{}, err
	}
	defer tx.Rollback()

	// remove existing cover
	_, err = tx.Exec(`
		DELETE FROM project_images
		WHERE projectId = ? AND type = 'cover'
	`, projectId)
	if err != nil {
		return types.ProjectImage{}, err
	}

	// insert new cover
	result, err := tx.Exec(`
		INSERT INTO project_images (projectId, url, type)
		VALUES (?, ?, 'cover')
	`, projectId, img.URL)
	if err != nil {
		return types.ProjectImage{}, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return types.ProjectImage{}, err
	}

	if err := tx.Commit(); err != nil {
		return types.ProjectImage{}, err
	}

	return s.GetProjectImageByID(int(id))
}

func (s *Store) UpdateProjectImagePositions(projectId int, ids []int) error {
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
