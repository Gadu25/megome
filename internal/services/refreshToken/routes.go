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
	// fmt.Println("=== HEADERS HANDLE REFRESH ===")
	// for key, values := range r.Header {
	// 	for _, value := range values {
	// 		fmt.Printf("%s: %s\n", key, value)
	// 	}
	// }
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("Error getting refresh cookie %v", err))
		return
	}

	newRefreshToken, newAccessToken, err := h.refreshStore.RefreshRotation(cookie.Value)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("Cookie error %v", err))
		return
	}

	utils.SetRefreshTokenCookie(w, newRefreshToken)
	utils.SetAccessTokenCookie(w, newAccessToken)

	resp := types.AuthResponse{
		Success: true,
		Message: "Token refreshed!",
	}
	utils.WriteJSON(w, http.StatusOK, resp)
}
