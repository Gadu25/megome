package utils

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"megome/config"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

var Validate = validator.New()

func ParseJSON(r *http.Request, payload any) error {
	if r.Body == nil {
		return fmt.Errorf("missing request body")
	}
	return json.NewDecoder(r.Body).Decode(payload)
}

func WriteJSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}

func WriteError(w http.ResponseWriter, status int, err error) {
	WriteJSON(w, status, map[string]string{"error": err.Error()})
}

func GetRequestId(r *http.Request) (int, error) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	return strconv.Atoi(idStr)
}

func SetRefreshTokenCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name: "refresh_token",
		// Domain: ".megome.com",
		Value:    token,
		HttpOnly: true,
		Secure:   true, // false only in local dev
		Path:     "/",
		SameSite: http.SameSiteNoneMode,
		Expires:  time.Now().Add(14 * 24 * time.Hour),
	})
}

func SetAccessTokenCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name: "Authentication",
		// Domain: ".megome.com",
		Value:    token,
		HttpOnly: true,
		Secure:   true, // false only in local dev
		Path:     "/",
		SameSite: http.SameSiteNoneMode,
	})
}

func ClearRefreshToken(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1, // immediately expire
		HttpOnly: true,
		Secure:   true, // set to true in production
		SameSite: http.SameSiteStrictMode,
	})
}

func ClearAccessToken(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "Authentication",
		Value:    "",
		Path:     "/",
		MaxAge:   -1, // immediately expire
		HttpOnly: true,
		Secure:   true, // set to true in production
		SameSite: http.SameSiteStrictMode,
	})
}

func ClearAllTokens(w http.ResponseWriter) {
	ClearAccessToken(w)
	ClearRefreshToken(w)
}

func GetFiletypeExtension(fileType string) (string, error) {
	switch fileType {
	case "image/jpeg":
		return ".jpg", nil
	case "image/png":
		return ".png", nil
	case "image/webp":
		return ".webp", nil
	default:
		return "", fmt.Errorf("unsupported file type: %s", fileType)
	}
}

func GetPublicFile(path string) string {
	baseUrl := config.Envs.R2PublicUrl
	return fmt.Sprintf("%s/%s", baseUrl, path)
}

func ExtractR2Key(input string) string {
	if input == "" {
		return ""
	}

	// Case 1: already a raw key (no scheme, no host)
	if !strings.Contains(input, "://") {
		return strings.TrimPrefix(input, "/")
	}

	// Case 2: full URL
	parsed, err := url.Parse(input)
	if err != nil {
		return ""
	}

	path := parsed.Path

	// remove leading slash to match R2 key format
	return strings.TrimPrefix(path, "/")
}

func GenerateUUID() string {
	return uuid.NewString()
}

func GenerateRandomToken(prefix string) (string, error) {
	b := make([]byte, 32) // 256-bit token

	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	token := hex.EncodeToString(b)

	return prefix + token, nil
}

func IsMysqlDuplicateKeyError(err error) bool {
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) {
		// reserved error number for duplicate entry
		return mysqlErr.Number == 1062
	}
	return false
}

func GetTokenFromRequest(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")

	parts := strings.SplitN(authHeader, " ", 2)

	if len(parts) != 2 || parts[0] != "Bearer" || parts[1] == "" {
		return ""
	}

	return parts[1]
}

func ParseIntOrDefault(value string, fallback int) int {
	if value == "" {
		return fallback
	}

	v, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}

	return v
}

func NilIfEmpty(s *string) *string {
	if s == nil || *s == "" {
		return nil
	}
	return s
}
