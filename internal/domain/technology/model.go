package technology

type TechnologyStore interface {
	GetTechnologies() ([]Technology, error)
	CreateTechnology(Technology) error
	UpdateTechnology(id int, technology Technology) error
	DeleteTechnology(id int) error
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
