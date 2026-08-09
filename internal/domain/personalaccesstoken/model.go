package personalaccesstoken

import "time"

type PersonalAccessTokenStore interface {
	GetPATByToken(string) (PATMinified, error)
	GetPATs(userId int, limit int, offset int) ([]PersonalAccessToken, error)
	CreatePAT(int, string) (string, error)
	RevokePAT(int, int) error
	DeletePAT(int, int) error
	GetTokenCountByUserID(int) (int, error)
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

type PaginatedPATResponse struct {
	Data       []PersonalAccessToken `json:"data"`
	Pagination struct {
		Limit  int `json:"limit"`
		Offset int `json:"offset"`
		Total  int `json:"total"`
	} `json:"pagination"`
}
