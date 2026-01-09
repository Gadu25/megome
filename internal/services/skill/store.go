package skill

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

func (s *Store) GetSkills(userID int) ([]types.Skill, error) {
	rows, err := s.db.Query(
		"SELECT * FROM skills WHERE userId = ?",
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var skills []types.Skill

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

func (s *Store) CreateSkill(skill types.Skill) error {
	_, err := s.db.Exec("INSERT INTO skills (userId, skillName, proficiency) VALUES (?, ?, ?)",
		skill.UserID,
		skill.SkillName,
		skill.Proficiency,
	)
	if err != nil {
		return err
	}
	return nil
}

func (s *Store) UpdateSkill(id int, skill types.Skill) error {
	_, err := s.db.Exec("UPDATE skills SET skillName = ?, proficiency = ?, updatedAt = CURRENT_TIMESTAMP WHERE id = ?",
		skill.SkillName,
		skill.Proficiency,
		id,
	)
	if err != nil {
		return err
	}
	return nil
}

func (s *Store) DeleteSkill(id int) error {
	_, err := s.db.Exec("DELETE FROM skills WHERE id = ?", id)
	if err != nil {
		return err
	}
	return nil
}

func scanRowIntoSkill(rows *sql.Rows) (types.Skill, error) {
	var skill types.Skill

	err := rows.Scan(
		&skill.ID,
		&skill.UserID,
		&skill.SkillName,
		&skill.Proficiency,
		&skill.CreatedAt,
		&skill.UpdatedAt,
	)
	if err != nil {
		return types.Skill{}, err
	}
	return skill, nil
}
