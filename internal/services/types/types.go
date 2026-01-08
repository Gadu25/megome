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
	UserID       int    `json:"userId`
	Bio          string `json:"bio"`
	Phone        string `json:"phone"`
	Website      string `json:"website"`
	Location     string `json:"location"`
	ProfileImage string `json:"profileImage"`
	CreatedAt    string `json:"createdAt`
	UpdatedAt    string `json:"updatedAt`
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
