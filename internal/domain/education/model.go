package education

type EducationStore interface {
	GetPublicEducations(userId int) ([]Education, error)
	GetEducations(userId int) ([]Education, error)
	CreateEducation(Education) (Education, error)
	UpdateEducation(id int, education Education) (Education, error)
	DeleteEducation(id int) (Education, error)
}

type Education struct {
	ID           int     `json:"id"`
	UserID       int     `json:"userId"`
	School       string  `json:"school"`
	Description  string  `json:"description"`
	Degree       string  `json:"degree"`
	FieldOfStudy string  `json:"fieldOfStudy"`
	StartDate    string  `json:"startDate"`
	EndDate      *string `json:"endDate"`
	IsPresent    bool    `json:"isPresent"`
	CreatedAt    string  `json:"createdAt"`
	UpdatedAt    string  `json:"updatedAt"`
}

type EducationPayload struct {
	School       string  `json:"school" validate:"required"`
	Description  string  `json:"description"`
	Degree       string  `json:"degree"`
	FieldOfStudy string  `json:"fieldOfStudy"`
	StartDate    string  `json:"startDate"`
	EndDate      *string `json:"endDate"`
	IsPresent    bool    `json:"isPresent"`
}
