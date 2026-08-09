package technology

type TechnologyStore interface {
	GetTechnologies(limit int, offset int) ([]Technology, error)
	CreateTechnology(Technology) error
	UpdateTechnology(id int, technology Technology) error
	DeleteTechnology(id int) error
	CountAll() (int, error)
}

type Technology struct {
	ID              int     `json:"id"`
	CreatedByUserId *int    `json:"createdByUserId"`
	Name            string  `json:"name"`
	Slug            string  `json:"slug"`
	Category        string  `json:"category"`
	IsVerified      string  `json:"isVerified"`
	CreatedAt       string  `json:"createdAt"`
	UpdatedAt       *string `json:"updatedAt"`
}

type TechnologyPayload struct {
	Name     string `json:"name"`
	Category string `json:"category"`
}

type PaginatedTechnologiesResponse struct {
	Data       []Technology `json:"data"`
	Pagination struct {
		Limit  int `json:"limit"`
		Offset int `json:"offset"`
		Total  int `json:"total"`
	} `json:"pagination"`
}
