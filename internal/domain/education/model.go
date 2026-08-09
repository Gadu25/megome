package education

type EducationStore interface {
	GetEducations(userId int, limit int, offset int) ([]Education, error)
	CreateEducation(Education) (Education, error)
	UpdateEducation(id int, education Education) (Education, error)
	DeleteEducation(id int, deletedBy int) (Education, error)
	ReorderEducations(userID int, items []ReorderItem) error
	CountByUserID(userId int) (int, error)
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
	DisplayOrder int     `json:"displayOrder"`
	CreatedAt    string  `json:"createdAt"`
	UpdatedAt    string  `json:"updatedAt"`
	DeletedAt    *string `json:"deletedAt,omitempty"`
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

type ReorderItem struct {
	ID           int `json:"id"`
	DisplayOrder int `json:"displayOrder"`
}

type PaginatedEducationsResponse struct {
	Data       []Education `json:"data"`
	Pagination struct {
		Limit  int `json:"limit"`
		Offset int `json:"offset"`
		Total  int `json:"total"`
	} `json:"pagination"`
}
