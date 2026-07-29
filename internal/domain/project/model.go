package project

import "megome/internal/domain/technology"

type ProjectStore interface {
	GetProjectById(int) (ProjectFull, error)
	GetProjects(int) ([]Project, error)
	GetProjectsFull(int) ([]ProjectFull, error)
	CreateProject(Project) (ProjectFull, error)
	UpdateProject(int, Project) (ProjectFull, error)
	DeleteProject(int) (ProjectFull, error)
	ReorderProjects(int, []ReorderItem) error
}

type Project struct {
	ID           int     `json:"id"`
	UserID       int     `json:"userId"`
	Title        string  `json:"title"`
	Description  string  `json:"description"`
	Link         string  `json:"link"`
	GithubLink   string  `json:"githubLink"`
	Status       string  `json:"status"`
	IsDraft      bool    `json:"isDraft"`
	DisplayOrder int     `json:"displayOrder"`
	CreatedAt    string  `json:"createdAt"`
	UpdatedAt    *string `json:"updatedAt"`
}

type ReorderItem struct {
	ID           int `json:"id"`
	DisplayOrder int `json:"displayOrder"`
}

type ProjectImages struct {
	Cover       *string  `json:"cover"`
	Screenshots []string `json:"screenshots"`
}

type ProjectFull struct {
	Project

	Images       ProjectImages          `json:"images"`
	Technologies []technology.Technology `json:"technologies"`
}

type ProjectPayload struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Link        string `json:"link"`
	GithubLink  string `json:"githubLink"`
	Status      string `json:"status"`
	IsDraft     bool   `json:"isDraft"`
}

type ProjectImageStore interface {
	GetProjectImageByID(int) (ProjectImage, error)
	GetProjectImages(int) ([]ProjectImage, error)
	AddProjectImage(ProjectImage) (ProjectImage, error)
	DeleteProjectImage(int) error
	SetProjectCover(int, ProjectImage) (ProjectImage, error)
}

type ProjectImage struct {
	ID        int    `json:"id"`
	ProjectID int    `json:"projectId"`
	URL       string `json:"url"`
	Type      string `json:"type"`
	Position  *int   `json:"position"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

type ProjectTechStore interface {
	CreateProjectTech(ProjectTech) error
	CreateProjectTechBatch(int, []int) error
	DeleteProjectTech(int) error
}

type ProjectTech struct {
	ID        int    `json:"id"`
	ProjectID int    `json:"projectId"`
	TechID    int    `json:"techId"`
	CreatedAt string `json:"createdAt"`
}

type ProjectTechPayload struct {
	ProjectID int `json:"projectId"`
	TechID    int `json:"techId"`
}

type BatchProjectTechPayload struct {
	TechIDs []int `json:"techIds" validate:"required,min=1"`
}
