package refreshToken

import (
	"fmt"
	"megome/internal/services/types"
	"megome/internal/services/utils"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

type Handler struct {
	refreshStore types.RefreshTokenStore
}

func NewHandler(refreshStore types.RefreshTokenStore) *Handler {
	return &Handler{refreshStore: refreshStore}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/auth/refresh", h.handleRefresh).Methods("POST")
}

func (h *Handler) handleRefresh(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("Error getting cookie %v", err))
		return
	}

	newRefreshToken, newAccessToken, err := h.refreshStore.RefreshRotation(cookie.Value)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("Cookie error %v", err))
		return
	}

	// Set refresh token as HttpOnly cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    newRefreshToken,
		HttpOnly: true,
		Secure:   true, // false only in local dev
		Path:     "/api/v1/auth/refresh",
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Now().Add(14 * 24 * time.Hour),
	})

	utils.WriteJSON(w, http.StatusOK, map[string]string{
		"message":      "Token refreshed!",
		"access-token": newAccessToken,
	})
}
