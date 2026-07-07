package refreshtoken

import (
	"database/sql"
	"time"
)

type RefreshTokenStore interface {
	CreateRefreshToken(userId int) (string, error)
	RefreshRotation(token string) (string, string, error)
	LogoutUser(token string) error
}

type RefreshToken struct {
	ID        int          `json:"id"`
	UserId    int          `json:"userId"`
	TokenHash string       `json:"tokenHash"`
	ExpiresAt time.Time    `json:"expiresAt"`
	RevokedAt sql.NullTime `json:"revokedAt"`
	CreatedAt string       `json:"createdAt"`
	UpdatedAt string       `json:"updatedAt"`
}

type AuthResponse struct {
	Success      bool   `json:"success"`
	Message      string `json:"message"`
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}
