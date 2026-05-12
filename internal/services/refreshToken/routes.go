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
	router.HandleFunc("/auth/refresh", h.handleRefresh).Methods("GET")
}

func (h *Handler) handleRefresh(w http.ResponseWriter, r *http.Request) {
	refreshToken := utils.GetTokenFromRequest(r)
	if refreshToken == "" {
		permissionDenied(w, "invalid token")
		return
	}

	newRefreshToken, newAccessToken, err := h.refreshStore.RefreshRotation(refreshToken)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("Cookie error %v", err))
		return
	}

	resp := types.AuthResponse{
		Message:      "Token refreshed!",
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
	}
	utils.WriteJSON(w, http.StatusOK, resp)
}

func permissionDenied(w http.ResponseWriter, m string) {
	utils.WriteError(w, http.StatusUnauthorized, fmt.Errorf("permission denied %v", m))
}
