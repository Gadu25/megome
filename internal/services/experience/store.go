package experience

import (
	"database/sql"
	"fmt"
	"megome/internal/services/types"
	"megome/internal/services/utils"
	"strings"
)

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) GetExperienceById(id int) (types.Experience, error) {
	row := s.db.QueryRow("SELECT id, userId, title, company, logo, startDate, endDate, isPresent, description, createdAt, updatedAt FROM experiences WHERE id = ?", id)

	var experience types.Experience
	err := row.Scan(
		&experience.ID,
		&experience.UserID,
		&experience.Title,
		&experience.Company,
		&experience.Logo,
		&experience.StartDate,
		&experience.EndDate,
		&experience.IsPresent,
		&experience.Description,
		&experience.CreatedAt,
		&experience.UpdatedAt,
	)

	if err != nil {
		return types.Experience{}, err
	}

	techsMap, err := s.GetExperienceTechs([]int{id})
	if err != nil {
		return types.Experience{}, err
	}

	experience.Technologies = techsMap[id]

	return experience, nil
}

func (s *Store) getExperiences(userID int) ([]types.Experience, error) {
	rows, err := s.db.Query(
		"SELECT id, userId, title, company, logo, startDate, endDate, isPresent, description, createdAt, updatedAt FROM experiences WHERE userId = ?",
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	experiences := make([]types.Experience, 0)

	for rows.Next() {
		exp, err := scanRowIntoExperience(rows)
		if err != nil {
			return nil, err
		}
		experiences = append(experiences, exp)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return experiences, nil
}

func (s *Store) populateTechs(experiences []types.Experience) ([]types.Experience, error) {
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

	result := make([]types.Experience, len(experiences))
	for i, exp := range experiences {
		exp.Technologies = techsMap[exp.ID]
		result[i] = exp
	}

	return result, nil
}

func (s *Store) GetPublicExperiences(userID int) ([]types.Experience, error) {
	experiences, err := s.getExperiences(userID)
	if err != nil {
		return nil, err
	}

	return s.populateTechs(experiences)
}

func (s *Store) GetExperiences(userID int) ([]types.Experience, error) {
	experiences, err := s.getExperiences(userID)
	if err != nil {
		return nil, err
	}

	return s.populateTechs(experiences)
}

func (s *Store) CreateExperience(experience types.Experience) (types.Experience, error) {
	result, err := s.db.Exec("INSERT INTO experiences (userId, title, company, logo, startDate, endDate, isPresent, description) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		experience.UserID,
		experience.Title,
		experience.Company,
		utils.NilIfEmpty(experience.Logo),
		experience.StartDate,
		utils.NilIfEmpty(experience.EndDate),
		experience.IsPresent,
		experience.Description,
	)
	if err != nil {
		return types.Experience{}, err
	}

	id, err := result.LastInsertId()

	if err != nil {
		return types.Experience{}, err
	}

	return s.GetExperienceById(int(id))
}

func (s *Store) UpdateExperience(id int, experience types.Experience) (types.Experience, error) {
	_, err := s.db.Exec("UPDATE experiences SET title = ?, company = ?, logo = ?, startDate = ?, endDate = ?, isPresent = ?, description = ?, updatedAt = CURRENT_TIMESTAMP WHERE id = ?",
		experience.Title,
		experience.Company,
		utils.NilIfEmpty(experience.Logo),
		experience.StartDate,
		utils.NilIfEmpty(experience.EndDate),
		experience.IsPresent,
		experience.Description,
		id,
	)
	if err != nil {
		return types.Experience{}, err
	}

	return s.GetExperienceById(id)
}

func (s *Store) DeleteExperience(id int) (types.Experience, error) {
	cert, err := s.GetExperienceById(id)

	if err != nil {
		return types.Experience{}, err
	}

	_, err = s.db.Exec("DELETE FROM experiences WHERE id = ?", id)
	if err != nil {
		return types.Experience{}, err
	}
	return cert, nil
}

func (s *Store) GetExperienceTechs(experienceIDs []int) (map[int][]types.Technology, error) {
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

	result := make(map[int][]types.Technology)

	for rows.Next() {
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
		if err != nil {
			return nil, err
		}

		result[experienceID] = append(result[experienceID], tech)
	}

	return result, rows.Err()
}

func scanRowIntoExperience(rows *sql.Rows) (types.Experience, error) {
	var experience types.Experience

	err := rows.Scan(
		&experience.ID,
		&experience.UserID,
		&experience.Title,
		&experience.Company,
		&experience.Logo,
		&experience.StartDate,
		&experience.EndDate,
		&experience.IsPresent,
		&experience.Description,
		&experience.CreatedAt,
		&experience.UpdatedAt,
	)
	if err != nil {
		return types.Experience{}, err
	}

	if experience.Logo != nil && *experience.Logo != "" {
		isFullURL := strings.HasPrefix(*experience.Logo, "https://") || strings.HasPrefix(*experience.Logo, "http://")

		if !isFullURL {
			*experience.Logo = utils.GetPublicFile(*experience.Logo)
		}
	}

	return experience, nil
}
