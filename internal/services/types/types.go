package types

import (
	"database/sql"
	"time"
)

type RefreshToken struct {
	ID        int          `json:"id"`
	UserId    int          `json:"userId"`
	TokenHash string       `json:"tokenHash"`
	ExpiresAt time.Time    `json:"expiresAt"`
	RevokedAt sql.NullTime `json:"revokedAt"`
	CreatedAt string       `json:"createdAt"`
	UpdatedAt string       `json:"updatedAt"`
}

type RefreshTokenStore interface {
	CreateRefreshToken(userId int) (string, error)
	RefreshRotation(token string) (string, string, error)
	LogoutUser(token string) error
}

type UserStore interface {
	GetUserByEmail(email string) (*User, error)
	GetUserByEmailOrUsername(input string) (*User, error)
	GetUserByID(id int) (*User, error)
	CreateUser(User) (*User, error)
	GetOAuthAccount(provider string, providerUserID string) (*OAuthAccount, error)
	CreateOAuthAccount(account OAuthAccount) error
}

type User struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Password  string    `json:"password"`
	CreatedAt time.Time `json:"createdAt"`
}

type RegisterUserPayload struct {
	Username string `json:"username" validate:"required,min=5,max=130"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=3,max=130"`
}

type LoginUserPayload struct {
	EmailOrUsername string `json:"emailOrUsername" validate:"required,min=3,max=130"`
	Password        string `json:"password" validate:"required"`
}

type AuthResponse struct {
	Success      bool   `json:"success"`
	Message      string `json:"message"`
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

type OAuthAccount struct {
	ID             int
	UserID         int
	Provider       string
	ProviderUserID string
	Email          *string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type GoogleUser struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
	Locale        string `json:"locale"`
}

type ProfileStore interface {
	GetPublicProfile(int) (*Profile, error)
	GetProfile(userId int) (*Profile, error)
	MakeProfile(Profile) error
	UpsertOAuthProfile(Profile) error
}

type Profile struct {
	ID           int    `json:"id"`
	UserID       int    `json:"userId"`
	FirstName    string `json:"firstName"`
	LastName     string `json:"lastName"`
	Title        string `json:"title"`
	Birthday     string `json:"birthday"`
	Bio          string `json:"bio"`
	Phone        string `json:"phone"`
	Website      string `json:"website"`
	Location     string `json:"location"`
	ProfileImage string `json:"profileImage"`
	CreatedAt    string `json:"createdAt"`
	UpdatedAt    string `json:"updatedAt"`
}
type PublicProfileResponse struct {
	ID           int    `json:"id"`
	FirstName    string `json:"firstName"`
	LastName     string `json:"lastName"`
	Title        string `json:"title"`
	Birthday     string `json:"birthday"`
	Bio          string `json:"bio"`
	Phone        string `json:"phone"`
	Website      string `json:"website"`
	Location     string `json:"location"`
	ProfileImage string `json:"profileImage"`
}

type MakeProfilePayload struct {
	Bio       string `json:"bio"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Birthday  string `json:"birthday"`
	Title     string `json:"title"`
	Phone     string `json:"phone"`
	Website   string `json:"website"`
	Location  string `json:"location"`
}

type ExperienceStore interface {
	GetPublicExperiences(userId int) ([]Experience, error)
	GetExperiences(userId int) ([]Experience, error)
	CreateExperience(Experience) (Experience, error)
	UpdateExperience(id int, Experience Experience) (Experience, error)
	DeleteExperience(id int) (Experience, error)
}

type Experience struct {
	ID          int     `json:"id"`
	UserID      int     `json:"userId"`
	Title       string  `json:"title"`
	Company     string  `json:"company"`
	StartDate   string  `json:"startDate"`
	EndDate     *string `json:"endDate"`
	IsPresent   bool    `json:"isPresent"`
	Description string  `json:"description"`
	CreatedAt   string  `json:"createdAt"`
	UpdatedAt   string  `json:"updatedAt"`
}

type ExperiencePayload struct {
	Title       string  `json:"title" validate:"required"`
	Company     string  `json:"company" validate:"required"`
	StartDate   string  `json:"startDate" validate:"required"`
	EndDate     *string `json:"endDate"`
	IsPresent   bool    `json:"isPresent"`
	Description string  `json:"description"`
}

type SkillStore interface {
	GetPublicSkills(userId int) ([]Skill, error)
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

type CertificationStore interface {
	GetPublicCertifications(userId int) ([]Certification, error)
	GetCertifications(userId int) ([]Certification, error)
	CreateCertification(Certification) (Certification, error)
	UpdateCertification(id int, certification Certification) (Certification, error)
	DeleteCertification(id int) (Certification, error)
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

type ProjectStore interface {
	GetPublicProjectById(int) (ProjectFull, error)
	GetProjectById(int) (ProjectFull, error)
	GetPublicProjects(int) ([]ProjectFull, error)
	GetProjects(int) ([]Project, error)
	GetProjectsFull(int) ([]ProjectFull, error)
	CreateProject(Project) (ProjectFull, error)
	UpdateProject(int, Project) (ProjectFull, error)
	DeleteProject(int) (ProjectFull, error)
}

type Project struct {
	ID          int     `json:"id"`
	UserID      int     `json:"userId"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Link        string  `json:"link"`
	GithubLink  string  `json:"githubLink"`
	Status      string  `json:"status"`
	IsDraft     bool    `json:"isDraft"`
	CreatedAt   string  `json:"createdAt"`
	UpdatedAt   *string `json:"updatedAt"`
}

type ProjectImages struct {
	Cover       *string  `json:"cover"`
	Screenshots []string `json:"screenshots"`
}

type ProjectFull struct {
	Project

	Images       ProjectImages `json:"images"`
	Technologies []Technology  `json:"technologies"`
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

type BatchProjectTechPayload struct {
	TechIDs []int `json:"techIds" validate:"required,min=1"`
}

type PersonalAccessToken struct {
	ID         int        `json:"id"`
	UserID     int        `json:"userId"`
	Name       string     `json:"name"`
	TokenHash  string     `json:"tokenHash"`
	LastUsedAt *time.Time `json:"lastUsedAt,omitempty"`
	RevokedAt  *time.Time `json:"revokedAt,omitempty"`
	CreatedAt  time.Time  `json:"createdAt"`
	UpdatedAt  time.Time  `json:"updatedAt"`
}

type PATMinified struct {
	ID        int
	UserID    int
	Name      string
	TokenHash string
	RevokedAt *time.Time
}

type PersonalAccessTokenPayload struct {
	Name string `json:"name" validate:"required"`
}

type APIUsageLog struct {
	ID             int       `json:"id"`
	UserID         int       `json:"userId"`
	TokenID        int       `json:"tokenId"`
	Endpoint       string    `json:"endpoint"`
	Method         string    `json:"method"`
	StatusCode     int       `json:"statusCode"`
	IPAddress      string    `json:"ipAddress,omitempty"`
	UserAgent      string    `json:"userAgent,omitempty"`
	ResponseTimeMs int       `json:"responseTimeMs,omitempty"`
	CreatedAt      time.Time `json:"createdAt"`
}

type APIUsageLogWithToken struct {
	Token PersonalAccessToken `json:"token"`
	Logs  []APIUsageLog       `json:"logs"`
}

type APIUsageLogStore interface {
	Create(log APIUsageLog) error
	GetByTokenID(tokenId int, limit int, offset int) (APIUsageLogWithToken, error)
}

type PersonalAccessTokenStore interface {
	GetPATByToken(string) (PATMinified, error)
	GetPATs(int) ([]PersonalAccessToken, error)
	CreatePAT(int, string) (string, error)
	RevokePAT(int, int) error
	DeletePAT(int, int) error
}
