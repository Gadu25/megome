package skill

type SkillStore interface {
	GetSkills(userId int, limit int, offset int) ([]Skill, error)
	CreateSkill(Skill) (Skill, error)
	UpdateSkill(id int, Skill Skill) (Skill, error)
	DeleteSkill(id int, deletedBy int) (Skill, error)
	CountByUserID(userId int) (int, error)
}

type Skill struct {
	ID          int    `json:"id"`
	UserID      int    `json:"userId"`
	SkillName   string `json:"skillName"`
	Proficiency string `json:"proficiency"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string  `json:"updatedAt"`
	DeletedAt   *string `json:"deletedAt,omitempty"`
}

type SkillPayload struct {
	SkillName   string `json:"skillName" validate:"required"`
	Proficiency string `json:"proficiency"`
}

type PaginatedSkillsResponse struct {
	Data       []Skill `json:"data"`
	Pagination struct {
		Limit  int `json:"limit"`
		Offset int `json:"offset"`
		Total  int `json:"total"`
	} `json:"pagination"`
}
