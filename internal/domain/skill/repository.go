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
	row := s.db.QueryRow("SELECT id, userId, skillName, proficiency, deletedAt, createdAt, updatedAt FROM skills WHERE id = ? AND deletedAt IS NULL", id)
	return scanRowIntoSkill(row)
}

func (s *Repository) GetSkills(userID int, limit int, offset int) ([]Skill, error) {
	rows, err := s.db.Query(
		"SELECT id, userId, skillName, proficiency, deletedAt, createdAt, updatedAt FROM skills WHERE userId = ? AND deletedAt IS NULL LIMIT ? OFFSET ?",
		userID,
		limit,
		offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanRowsIntoSkill(rows)
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

func (s *Repository) DeleteSkill(id int, deletedBy int) (Skill, error) {
	skill, err := s.GetSkillById(id)

	if err != nil {
		return Skill{}, err
	}

	_, err = s.db.Exec("UPDATE skills SET deletedAt = NOW() WHERE id = ? AND deletedAt IS NULL", id)
	if err != nil {
		return Skill{}, err
	}

	return skill, nil
}

func scanRowIntoSkill(scanner interface{ Scan(dest ...interface{}) error }) (Skill, error) {
	var skill Skill
	err := scanner.Scan(
		&skill.ID,
		&skill.UserID,
		&skill.SkillName,
		&skill.Proficiency,
		&skill.DeletedAt,
		&skill.CreatedAt,
		&skill.UpdatedAt,
	)
	if err != nil {
		return Skill{}, err
	}
	return skill, nil
}

func scanRowsIntoSkill(rows *sql.Rows) ([]Skill, error) {
	var skills []Skill
	for rows.Next() {
		skill, err := scanRowIntoSkill(rows)
		if err != nil {
			return nil, err
		}
		skills = append(skills, skill)
	}
	return skills, rows.Err()
}

func (s *Repository) CountByUserID(userID int) (int, error) {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM skills WHERE userId = ? AND deletedAt IS NULL", userID).Scan(&count)
	return count, err
}
