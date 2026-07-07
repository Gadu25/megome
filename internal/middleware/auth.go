package middleware

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"megome/internal/config"
	"megome/internal/domain/personalaccesstoken"
	"megome/internal/domain/user"
	"megome/internal/pkg/auth"
	"megome/internal/pkg/httputil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

func WithJWTAuth(handlerFunc http.HandlerFunc, store user.UserStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenString := httputil.GetTokenFromRequest(r)
		if tokenString == "" {
			permissionDenied(w, "invalid token")
			return
		}

		token, err := jwt.ParseWithClaims(tokenString, &auth.Claims{}, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(config.Envs.JWTSecret), nil
		})
		if err != nil {
			log.Printf("token validation failed: %v", err)
			permissionDenied(w, "invalid token")
			return
		}

		claims, ok := token.Claims.(*auth.Claims)
		if !ok || !token.Valid {
			permissionDenied(w, "invalid token")
			return
		}
		if claims.ExpiresAt == nil || claims.ExpiresAt.Time.Before(time.Now()) {
			log.Println("token expired")
			permissionDenied(w, "token expired")
			return
		}

		userID, err := strconv.Atoi(claims.UserID)
		u, err := store.GetUserByID(userID)
		if err != nil {
			log.Printf("failed to fetch user: %v", err)
			permissionDenied(w, "failed to fetch user")
			return
		}

		ctx := context.WithValue(r.Context(), auth.UserKey, u.ID)
		handlerFunc(w, r.WithContext(ctx))
	}
}

func permissionDenied(w http.ResponseWriter, m string) {
	httputil.WriteError(w, http.StatusUnauthorized, fmt.Errorf("permission denied %v", m))
}

func WithPATAuth(handlerFunc http.HandlerFunc, store personalaccesstoken.PersonalAccessTokenStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")

		if authHeader == "" {
			httputil.WriteError(w, http.StatusUnauthorized, errors.New("missing authorization header"))
			return
		}

		parts := strings.Split(authHeader, " ")

		if len(parts) != 2 || parts[0] != "Bearer" {
			httputil.WriteError(w, http.StatusUnauthorized, errors.New("invalid authorization format"))
			return
		}

		rawToken := parts[1]

		hash := sha256.Sum256([]byte(rawToken))
		tokenHash := hex.EncodeToString(hash[:])

		token, err := store.GetPATByToken(tokenHash)
		if err != nil {
			httputil.WriteError(w, http.StatusUnauthorized, errors.New("invalid token"))
			return
		}

		// pointer nil check
		if token.RevokedAt != nil {
			httputil.WriteError(w, http.StatusUnauthorized, errors.New("token revoked"))
			return
		}

		ctx := context.WithValue(r.Context(), auth.PATUserIDKey, token.UserID)
		ctx = context.WithValue(ctx, auth.PATTokenIDKey, token.ID)

		handlerFunc(w, r.WithContext(ctx))
	}
}

func GetUserIDFromContext(ctx context.Context) int {
	userID, ok := ctx.Value(auth.UserKey).(int)
	if !ok {
		return -1
	}

	return userID
}

func GetPATUserIDFromContext(ctx context.Context) int {
	userID, _ := ctx.Value(auth.PATUserIDKey).(int)
	return userID
}

func GetPATTokenIDFromContext(ctx context.Context) int {
	tokenID, _ := ctx.Value(auth.PATTokenIDKey).(int)
	return tokenID
}
