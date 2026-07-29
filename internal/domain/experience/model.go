package experience

import "megome/internal/domain/technology"

type ExperienceStore interface {
	GetExperiences(userId int) ([]Experience, error)
	GetExperienceById(id int) (Experience, error)
	CreateExperience(Experience) (Experience, error)
	UpdateExperience(id int, Experience Experience) (Experience, error)
	DeleteExperience(id int) (Experience, error)
	ReorderExperiences(userID int, items []ReorderItem) error
}

type Experience struct {
	ID           int                      `json:"id"`
	UserID       int                      `json:"userId"`
	Title        string                   `json:"title"`
	Company      string                   `json:"company"`
	Logo         *string                  `json:"logo"`
	StartDate    string                   `json:"startDate"`
	EndDate      *string                  `json:"endDate"`
	IsPresent    bool                     `json:"isPresent"`
	Description  string                   `json:"description"`
	Technologies []technology.Technology  `json:"technologies"`
	DisplayOrder int                      `json:"displayOrder"`
	CreatedAt    string                   `json:"createdAt"`
	UpdatedAt    string                   `json:"updatedAt"`
}

type ReorderItem struct {
	ID           int `json:"id"`
	DisplayOrder int `json:"displayOrder"`
}

type ExperiencePayload struct {
	Title       string  `json:"title" validate:"required"`
	Company     string  `json:"company" validate:"required"`
	Logo        *string `json:"logo"`
	StartDate   string  `json:"startDate" validate:"required"`
	EndDate     *string `json:"endDate"`
	IsPresent   bool    `json:"isPresent"`
	Description string  `json:"description"`
}

type ExperienceTechStore interface {
	CreateExperienceTech(ExperienceTech) error
	CreateExperienceTechBatch(int, []int) error
	DeleteExperienceTech(int) error
	GetExperienceTechs([]int) (map[int][]technology.Technology, error)
}

type ExperienceTech struct {
	ID           int    `json:"id"`
	ExperienceID int    `json:"experienceId"`
	TechID       int    `json:"techId"`
	CreatedAt    string `json:"createdAt"`
}

type ExperienceTechPayload struct {
	ExperienceID int `json:"experienceId"`
	TechID       int `json:"techId"`
}

type BatchExperienceTechPayload struct {
	TechIDs []int `json:"techIds" validate:"required,min=1"`
}
