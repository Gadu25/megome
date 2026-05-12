package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"strings"

	"megome/internal/services/types"
	"megome/internal/services/utils"
)

func WithPATAuth(handlerFunc http.HandlerFunc, store types.PersonalAccessTokenStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")

		if authHeader == "" {
			utils.WriteError(w, http.StatusUnauthorized, errors.New("missing authorization header"))
			return
		}

		parts := strings.Split(authHeader, " ")

		if len(parts) != 2 || parts[0] != "Bearer" {
			utils.WriteError(w, http.StatusUnauthorized, errors.New("invalid authorization format"))
			return
		}

		rawToken := parts[1]

		hash := sha256.Sum256([]byte(rawToken))
		tokenHash := hex.EncodeToString(hash[:])

		token, err := store.GetPATByToken(tokenHash)
		if err != nil {
			utils.WriteError(w, http.StatusUnauthorized, errors.New("invalid token"))
			return
		}

		// pointer nil check
		if token.RevokedAt != nil {
			utils.WriteError(w, http.StatusUnauthorized, errors.New("token revoked"))
			return
		}

		ctx := context.WithValue(r.Context(), PATUserIDKey, token.UserID)
		ctx = context.WithValue(ctx, PATTokenIDKey, token.ID)

		handlerFunc(w, r.WithContext(ctx))
	}
}

func GetPATUserIDFromContext(ctx context.Context) int {
	userID, _ := ctx.Value(PATUserIDKey).(int)
	return userID
}

func GetPATTokenIDFromContext(ctx context.Context) int {
	tokenID, _ := ctx.Value(PATTokenIDKey).(int)
	return tokenID
}
