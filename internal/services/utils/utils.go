package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-playground/validator/v10"
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
		Secure:   false, // false only in local dev
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(14 * 24 * time.Hour),
	})
}

func SetAccessTokenCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name: "Authentication",
		// Domain: ".megome.com",
		Value:    token,
		HttpOnly: true,
		Secure:   false, // false only in local dev
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
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