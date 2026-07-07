package auth

type contextKey string

const (
	UserKey contextKey = "userID"

	PATUserIDKey  contextKey = "patUserId"
	PATTokenIDKey contextKey = "patTokenId"
)
