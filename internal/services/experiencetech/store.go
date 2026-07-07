package experiencetech

import (
	"database/sql"
	"fmt"
	"megome/internal/services/types"
	"strings"
)

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) GetExperienceTechById(id int) (*types.ExperienceTech, error) {
	row := s.db.QueryRow("SELECT id, experienceId, techId FROM experience_techs WHERE id = ?", id)
	expTech := new(types.ExperienceTech)
	err := row.Scan(
		&expTech.ID,
		&expTech.ExperienceID,
		&expTech.TechID,
	)
	if err != nil {
		return nil, err
	}
	return expTech, nil
}

func (s *Store) CreateExperienceTech(expTech types.ExperienceTech) error {
	_, err := s.db.Exec(
		"INSERT INTO experience_techs (experienceId, techId) VALUES (?, ?)",
		expTech.ExperienceID,
		expTech.TechID,
	)
	if err != nil {
		return err
	}
	return nil
}

func (s *Store) CreateExperienceTechBatch(experienceID int, techIDs []int) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	stmt, err := tx.Prepare(`INSERT INTO experience_techs (experienceId, techId) VALUES (?, ?)`)
	if err != nil {
		tx.Rollback()
		return err
	}

	defer stmt.Close()

	for _, techID := range techIDs {
		_, err = stmt.Exec(experienceID, techID)
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

func (s *Store) DeleteExperienceTech(id int) error {
	_, err := s.GetExperienceTechById(id)
	if err != nil {
		return err
	}
	_, err = s.db.Exec("DELETE FROM experience_techs WHERE id = ?", id)
	if err != nil {
		return err
	}
	return nil
}

func (s *Store) GetExperienceTechs(experienceIDs []int) (map[int][]types.Technology, error) {
	query, args := buildInQuery(experienceIDs)
	if query == "" {
		return nil, nil
	}

	rows, err := s.db.Query(fmt.Sprintf(`
		SELECT
			et.experienceId,
			t.id,
			t.createdByUserId,
			t.name,
			t.slug,
			t.category,
			t.isVerified,
			t.createdAt,
			t.updatedAt
		FROM experience_techs et
		INNER JOIN technologies t ON et.techId = t.id
		WHERE et.experienceId IN (%s)
	`, query), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[int][]types.Technology)

	for rows.Next() {
		experienceID, tech, err := scanTechnology(rows)
		if err != nil {
			return nil, err
		}

		result[experienceID] = append(result[experienceID], tech)
	}

	return result, rows.Err()
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

func scanTechnology(rows *sql.Rows) (int, types.Technology, error) {
	var (
		experienceID int
		tech         types.Technology
	)

	err := rows.Scan(
		&experienceID,
		&tech.ID,
		&tech.CreatedByUserId,
		&tech.Name,
		&tech.Slug,
		&tech.Category,
		&tech.IsVerified,
		&tech.CreatedAt,
		&tech.UpdatedAt,
	)

	return experienceID, tech, err
}
