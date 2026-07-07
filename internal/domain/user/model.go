package user

import "time"

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

type ForgotPassChangePayload struct {
	Token    string `json:"token"`
	Password string `json:"password"`
}

type ForgotPassPayload struct {
	Email string `json:"email"`
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
