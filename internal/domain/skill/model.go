package skill

type SkillStore interface {
	GetSkills(userId int) ([]Skill, error)
	CreateSkill(Skill) (Skill, error)
	UpdateSkill(id int, Skill Skill) (Skill, error)
	DeleteSkill(id int) (Skill, error)
}

type Skill struct {
	ID          int    `json:"id"`
	UserID      int    `json:"userId"`
	SkillName   string `json:"skillName"`
	Proficiency string `json:"proficiency"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}

type SkillPayload struct {
	SkillName   string `json:"skillName" validate:"required"`
	Proficiency string `json:"proficiency"`
}
