package refreshtoken

import (
	"database/sql"
	"time"
)

const (
	// RefreshTokenLifetime is how long a refresh token remains valid.
	RefreshTokenLifetime = 14 * 24 * time.Hour
	// RefreshTokenMaxAgeSeconds is RefreshTokenLifetime expressed in seconds for cookies.
	RefreshTokenMaxAgeSeconds = int64(14 * 24 * 60 * 60)
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
	Success            bool   `json:"success"`
	Message            string `json:"message"`
	AccessToken        string `json:"accessToken"`
	RefreshToken       string `json:"refreshToken"`
	AccessTokenMaxAge  int64  `json:"accessTokenMaxAge,omitempty"`
	RefreshTokenMaxAge int64  `json:"refreshTokenMaxAge,omitempty"`
}

// NewAuthResponse builds a successful AuthResponse with the token lifetimes
// the frontend needs to set cookie maxAge correctly.
func NewAuthResponse(message, accessToken, refreshToken string, accessMaxAge int64) AuthResponse {
	return AuthResponse{
		Success:            true,
		Message:            message,
		AccessToken:        accessToken,
		RefreshToken:       refreshToken,
		AccessTokenMaxAge:  accessMaxAge,
		RefreshTokenMaxAge: RefreshTokenMaxAgeSeconds,
	}
}

type SessionInfo struct {
	ID        int       `json:"id"`
	UserId    int       `json:"userId"`
	ExpiresAt time.Time `json:"expiresAt"`
	CreatedAt time.Time `json:"createdAt"`
}
