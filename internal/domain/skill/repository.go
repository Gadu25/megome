package skill

import (
	"database/sql"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (s *Repository) GetSkillById(id int) (Skill, error) {
	row := s.db.QueryRow("SELECT * FROM skills WHERE id = ?", id)

	var skill Skill
	err := row.Scan(
		&skill.ID,
		&skill.UserID,
		&skill.SkillName,
		&skill.Proficiency,
		&skill.CreatedAt,
		&skill.UpdatedAt,
	)

	if err != nil {
		return Skill{}, err
	}

	return skill, nil
}

func (s *Repository) GetPublicSkills(userID int) ([]Skill, error) {
	rows, err := s.db.Query(
		"SELECT * FROM skills WHERE userId = ?",
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	skills := make([]Skill, 0)

	for rows.Next() {
		skill, err := scanRowIntoSkill(rows)
		if err != nil {
			return nil, err
		}
		skills = append(skills, skill)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return skills, nil
}

func (s *Repository) GetSkills(userID int) ([]Skill, error) {
	rows, err := s.db.Query(
		"SELECT * FROM skills WHERE userId = ?",
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	skills := make([]Skill, 0)

	for rows.Next() {
		skill, err := scanRowIntoSkill(rows)
		if err != nil {
			return nil, err
		}
		skills = append(skills, skill)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return skills, nil
}

func (s *Repository) CreateSkill(skill Skill) (Skill, error) {
	result, err := s.db.Exec("INSERT INTO skills (userId, skillName, proficiency) VALUES (?, ?, ?)",
		skill.UserID,
		skill.SkillName,
		skill.Proficiency,
	)

	if err != nil {
		return Skill{}, err
	}

	id, err := result.LastInsertId()

	if err != nil {
		return Skill{}, err
	}

	return s.GetSkillById(int(id))
}

func (s *Repository) UpdateSkill(id int, skill Skill) (Skill, error) {
	_, err := s.db.Exec("UPDATE skills SET skillName = ?, proficiency = ?, updatedAt = CURRENT_TIMESTAMP WHERE id = ?",
		skill.SkillName,
		skill.Proficiency,
		id,
	)
	if err != nil {
		return Skill{}, err
	}
	return s.GetSkillById(id)
}

func (s *Repository) DeleteSkill(id int) (Skill, error) {
	skill, err := s.GetSkillById(id)

	if err != nil {
		return Skill{}, err
	}

	_, err = s.db.Exec("DELETE FROM skills WHERE id = ?", id)
	if err != nil {
		return Skill{}, err
	}

	return skill, nil
}

func scanRowIntoSkill(rows *sql.Rows) (Skill, error) {
	var skill Skill

	err := rows.Scan(
		&skill.ID,
		&skill.UserID,
		&skill.SkillName,
		&skill.Proficiency,
		&skill.CreatedAt,
		&skill.UpdatedAt,
	)
	if err != nil {
		return Skill{}, err
	}
	return skill, nil
}
