package experience

import (
	"database/sql"
	"fmt"
	"megome/internal/domain/technology"
	"megome/internal/pkg/httputil"
	"strings"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (s *Repository) GetExperienceById(id int) (Experience, error) {
	row := s.db.QueryRow("SELECT id, userId, title, company, logo, startDate, endDate, isPresent, description, displayOrder, createdAt, updatedAt FROM experiences WHERE id = ?", id)
	experience, err := scanRowIntoExperience(row)
	if err != nil {
		return Experience{}, err
	}

	techsMap, err := s.GetExperienceTechs([]int{id})
	if err != nil {
		return Experience{}, err
	}

	experience.Technologies = techsMap[id]

	return experience, nil
}

func (s *Repository) getExperiences(userID int) ([]Experience, error) {
	rows, err := s.db.Query(
		"SELECT id, userId, title, company, logo, startDate, endDate, isPresent, description, displayOrder, createdAt, updatedAt FROM experiences WHERE userId = ?",
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanRowsIntoExperience(rows)
}

func (s *Repository) populateTechs(experiences []Experience) ([]Experience, error) {
	if len(experiences) == 0 {
		return experiences, nil
	}

	ids := make([]int, 0, len(experiences))
	for _, exp := range experiences {
		ids = append(ids, exp.ID)
	}

	techsMap, err := s.GetExperienceTechs(ids)
	if err != nil {
		return nil, err
	}

	result := make([]Experience, len(experiences))
	for i, exp := range experiences {
		exp.Technologies = techsMap[exp.ID]
		result[i] = exp
	}

	return result, nil
}

func (s *Repository) GetExperiences(userID int) ([]Experience, error) {
	experiences, err := s.getExperiences(userID)
	if err != nil {
		return nil, err
	}

	return s.populateTechs(experiences)
}

func (s *Repository) CreateExperience(experience Experience) (Experience, error) {
	result, err := s.db.Exec("INSERT INTO experiences (userId, title, company, logo, startDate, endDate, isPresent, description) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		experience.UserID,
		experience.Title,
		experience.Company,
		httputil.NilIfEmpty(experience.Logo),
		experience.StartDate,
		httputil.NilIfEmpty(experience.EndDate),
		experience.IsPresent,
		experience.Description,
	)
	if err != nil {
		return Experience{}, err
	}

	id, err := result.LastInsertId()

	if err != nil {
		return Experience{}, err
	}

	return s.GetExperienceById(int(id))
}

func (s *Repository) UpdateExperience(id int, experience Experience) (Experience, error) {
	_, err := s.db.Exec("UPDATE experiences SET title = ?, company = ?, logo = ?, startDate = ?, endDate = ?, isPresent = ?, description = ?, updatedAt = CURRENT_TIMESTAMP WHERE id = ?",
		experience.Title,
		experience.Company,
		httputil.NilIfEmpty(experience.Logo),
		experience.StartDate,
		httputil.NilIfEmpty(experience.EndDate),
		experience.IsPresent,
		experience.Description,
		id,
	)
	if err != nil {
		return Experience{}, err
	}

	return s.GetExperienceById(id)
}

func (s *Repository) DeleteExperience(id int) (Experience, error) {
	cert, err := s.GetExperienceById(id)

	if err != nil {
		return Experience{}, err
	}

	_, err = s.db.Exec("DELETE FROM experiences WHERE id = ?", id)
	if err != nil {
		return Experience{}, err
	}
	return cert, nil
}

func (s *Repository) GetExperienceTechs(experienceIDs []int) (map[int][]technology.Technology, error) {
	if len(experienceIDs) == 0 {
		return nil, nil
	}

	placeholders := make([]string, len(experienceIDs))
	args := make([]interface{}, len(experienceIDs))

	for i, id := range experienceIDs {
		placeholders[i] = "?"
		args[i] = id
	}

	query := fmt.Sprintf(`
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
	`, strings.Join(placeholders, ","))

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[int][]technology.Technology)

	for rows.Next() {
		var (
			experienceID int
			tech         technology.Technology
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
		if err != nil {
			return nil, err
		}

		result[experienceID] = append(result[experienceID], tech)
	}

	return result, rows.Err()
}

func scanRowIntoExperience(scanner interface{ Scan(dest ...interface{}) error }) (Experience, error) {
	var experience Experience
	err := scanner.Scan(
		&experience.ID,
		&experience.UserID,
		&experience.Title,
		&experience.Company,
		&experience.Logo,
		&experience.StartDate,
		&experience.EndDate,
		&experience.IsPresent,
		&experience.Description,
		&experience.DisplayOrder,
		&experience.CreatedAt,
		&experience.UpdatedAt,
	)
	if err != nil {
		return Experience{}, err
	}

	if experience.Logo != nil && *experience.Logo != "" {
		if !strings.HasPrefix(*experience.Logo, "https://") && !strings.HasPrefix(*experience.Logo, "http://") {
			*experience.Logo = httputil.GetPublicFile(*experience.Logo)
		}
	}

	return experience, nil
}

func scanRowsIntoExperience(rows *sql.Rows) ([]Experience, error) {
	var experiences []Experience
	for rows.Next() {
		exp, err := scanRowIntoExperience(rows)
		if err != nil {
			return nil, err
		}
		experiences = append(experiences, exp)
	}
	return experiences, rows.Err()
}

func (s *Repository) GetExperienceTechById(id int) (*ExperienceTech, error) {
	row := s.db.QueryRow("SELECT id, experienceId, techId FROM experience_techs WHERE id = ?", id)
	expTech := new(ExperienceTech)
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

func (s *Repository) CreateExperienceTech(expTech ExperienceTech) error {
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

func (s *Repository) CreateExperienceTechBatch(experienceID int, techIDs []int) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	_, err = tx.Exec(`DELETE FROM experience_techs WHERE experienceId = ?`, experienceID)
	if err != nil {
		tx.Rollback()
		return err
	}

	if len(techIDs) == 0 {
		if err = tx.Commit(); err != nil {
			return err
		}
		return nil
	}

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

func (s *Repository) DeleteExperienceTech(id int) error {
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

func (s *Repository) ReorderExperiences(userID int, items []ReorderItem) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare("UPDATE experiences SET displayOrder = ?, updatedAt = CURRENT_TIMESTAMP WHERE id = ? AND userId = ?")
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
