package refreshToken

import (
	"fmt"
	"megome/internal/services/types"
	"megome/internal/services/utils"
	"net/http"

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

	utils.SetRefreshTokenCookie(w, newRefreshToken)

	utils.WriteJSON(w, http.StatusOK, map[string]string{
		"message":      "Token refreshed!",
		"access-token": newAccessToken,
	})
}
