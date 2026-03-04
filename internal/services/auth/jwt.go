package auth

import (
	"context"
	"fmt"
	"log"
	"megome/config"
	"megome/internal/services/types"
	"megome/internal/services/utils"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type contextKey string

type Claims struct {
	UserID string `json:"userId"`
	jwt.RegisteredClaims
}

const UserKey contextKey = "userID"

// func CreateJWT(secret []byte, userID int) (string, error) {
// 	expiration := time.Second * time.Duration(config.Envs.JWTExpirationInSeconds)
// 	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
// 		"userId":    strconv.Itoa(userID),
// 		"expiresAt": time.Now().Add(expiration).Unix(),
// 	})

// 	tokenString, err := token.SignedString(secret)
// 	if err != nil {
// 		return "", err
// 	}

// 	return tokenString, nil
// }

func CreateJWT(secret []byte, userID int) (string, error) {
	expiration := time.Duration(config.Envs.JWTExpirationInSeconds) * time.Second

	claims := Claims{
		UserID: strconv.Itoa(userID),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString(secret)
}

func WithJWTAuth(handlerFunc http.HandlerFunc, store types.UserStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenString := getTokenFromRequest(r)
		if tokenString == "" {
			permissionDenied(w, "invalid token")
			return
		}

		claims, err := validateToken(tokenString)
		if err != nil {
			log.Printf("token validation failed: %v", err)
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

		ctx := context.WithValue(r.Context(), UserKey, u.ID)
		handlerFunc(w, r.WithContext(ctx))
	}
}

func getTokenFromRequest(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return ""
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return ""
	}

	return parts[1]
}

func validateToken(t string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(t, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(config.Envs.JWTSecret), nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}

func permissionDenied(w http.ResponseWriter, m string) {
	utils.WriteError(w, http.StatusForbidden, fmt.Errorf("permission denied %v", m))
}

func GetUserIDFromContext(ctx context.Context) int {
	userID, ok := ctx.Value(UserKey).(int)
	if !ok {
		return -1
	}

	return userID
}
