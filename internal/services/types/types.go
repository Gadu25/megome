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
	GetProfile(userId int) ([]Profile, error)
}

type Profile struct {
	ID           int    `json:"id"`
	UserID       int    `json:"userId"`
	Bio          string `json:"bio"`
	Phone        string `json:"phone"`
	Website      string `json:"website"`
	Location     string `json:"location"`
	ProfileImage string `json:"profileImage"`
	CreatedAt    string `json:"createdAt`
	UpdatedAt    string `json:"updatedAt`
}
