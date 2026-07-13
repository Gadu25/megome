package apilog

import (
	"megome/internal/domain/personalaccesstoken"
	"time"
)

type APIUsageLogStore interface {
	Create(log APIUsageLog) error
	GetByTokenID(tokenId int, limit int, offset int) (APIUsageLogWithToken, error)
	GetUserUsageStats(userId int) (UserAPIUsageStats, error)
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
	Token personalaccesstoken.PersonalAccessToken `json:"token"`
	Logs  []APIUsageLog                           `json:"logs"`
}

type UserAPIUsageStats struct {
	RequestCount      int     `json:"requestCount"`
	AverageResponseMs float64 `json:"averageResponseMs"`
}
