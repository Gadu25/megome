package types

import (
	"time"
)

type UserStore interface {
	GetUserByEmail(email string) (*User, error)
	GetUserByID(id int) (*User, error)
	CreateUser(User) error
}

type User struct {
	ID        int       `json:"id"`
	FirstName string    `json:"firstName"`
	LastName  string    `json:"lastName"`
	Email     string    `json:"email"`
	Password  string    `json:"password"`
	CreatedAt time.Time `json:"createdAt"`
}

type RegisterUserPayload struct {
	FirstName string `json:"firstName" validate:"required"`
	LastName  string `json:"lastName" validate:"required"`
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required,min=3,max=130"`
}

type LoginUserPayload struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type ProfileStore interface {
	GetProfile(userId int) (*Profile, error)
	MakeProfile(Profile) error
}

type Profile struct {
	ID           int    `json:"id"`
	UserID       int    `json:"userId"`
	Bio          string `json:"bio"`
	Phone        string `json:"phone"`
	Website      string `json:"website"`
	Location     string `json:"location"`
	ProfileImage string `json:"profileImage"`
	CreatedAt    string `json:"createdAt"`
	UpdatedAt    string `json:"updatedAt"`
}

type MakeProfilePayload struct {
	Bio          string `json:"bio"`
	Phone        string `json:"phone"`
	Website      string `json:"website"`
	Location     string `json:"location"`
	ProfileImage string `json:"profileImage"`
}

type ExperienceStore interface {
	GetExperiences(userId int) ([]Experience, error)
	CreateExperience(Experience) error
	UpdateExperience(id int, Experience Experience) error
	DeleteExperience(id int) error
}

type Experience struct {
	ID          int     `json:"id"`
	UserID      int     `json:"userId"`
	Title       string  `json:"title"`
	Company     string  `json:"company"`
	StartDate   string  `json:"startDate"`
	EndDate     *string `json:"endDate"`
	Description string  `json:"description"`
	CreatedAt   string  `json:"createdAt"`
	UpdatedAt   string  `json:"updatedAt"`
}

type ExperiencePayload struct {
	Title       string  `json:"title" validate:"required"`
	Company     string  `json:"company" validate:"required"`
	StartDate   string  `json:"startDate" validate:"required"`
	EndDate     *string `json:"endDate"`
	Description string  `json:"description"`
}

type SkillStore interface {
	GetSkills(userId int) ([]Skill, error)
	CreateSkill(Skill) error
	UpdateSkill(id int, Skill Skill) error
	DeleteSkill(id int) error
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

type EducationStore interface {
	GetEducations(userId int) ([]Education, error)
	CreateEducation(Education) error
	UpdateEducation(id int, education Education) error
	DeleteEducation(id int) error
}

type Education struct {
	ID           int    `json:"id"`
	UserID       int    `json:"userId"`
	School       string `json:"school"`
	Degree       string `json:"degree"`
	FieldOfStudy string `json:"fieldOfStudy"`
	StartDate    string `json:"startDate"`
	EndDate      string `json:"endDate"`
	CreatedAt    string `json:"createdAt"`
	UpdatedAt    string `json:"updatedAt"`
}

type EducationPayload struct {
	School       string `json:"school" validate:"required"`
	Degree       string `json:"degree"`
	FieldOfStudy string `json:"fieldOfStudy"`
	StartDate    string `json:"startDate"`
	EndDate      string `json:"endDate"`
}

type CertificationStore interface {
	GetCertifications(userId int) ([]Certification, error)
	CreateCertification(Certification) error
	UpdateCertification(id int, certification Certification) error
	DeleteCertification(id int) error
}

type Certification struct {
	ID             int     `json:"id"`
	UserID         int     `json:"userId"`
	Title          string  `json:"title"`
	Issuer         string  `json:"issuer"`
	IssueDate      string  `json:"issueDate"`
	ExpirationDate *string `json:"expirationDate"`
	CredentialId   *string `json:"credentialId"`
	CredentialUrl  *string `json:"credentialUrl"`
	CreatedAt      string  `json:"createdAt"`
	UpdatedAt      string  `json:"updatedAt"`
}

type CertificationPayload struct {
	Title          string  `json:"title"`
	Issuer         string  `json:"issuer"`
	IssueDate      string  `json:"issueDate"`
	ExpirationDate *string `json:"expirationDate"`
	CredentialId   *string `json:"credentialId"`
	CredentialUrl  *string `json:"credentialUrl"`
}

type TechnologyStore interface {
	GetTechnologies(userId int) ([]Technology, error)
	CreateTechnology(Technology) error
	UpdateTechnology(id int, technology Technology) error
	DeleteTechnology(id int) error
}

type Technology struct {
	ID        int     `json:"id"`
	UserID    int     `json:"userId"`
	Name      string  `json:"name"`
	Slug      string  `json:"slug"`
	CreatedAt string  `json:"createdAt"`
	UpdatedAt *string `json:"updatedAt"`
}

type TechnologyPayload struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type ProjectStore interface {
	GetProjects(int) ([]Project, error)
	CreateProject(Project) error
	UpdateProject(int, Project) error
	DeleteProject(int) error
}

type Project struct {
	ID          int    `json:"id"`
	UserID      int    `json:"userId"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Link        string `json:"link"`
	GithubLink  string `json:"githubLink"`
}

type ProjectPayload struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Link        string `json:"link"`
	GithubLink  string `json:"githubLink"`
}

type ProjectTechStore interface {
	CreateProjectTech(ProjectTech) error
	DelteProjectTech(int) error
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
