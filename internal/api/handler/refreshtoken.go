package handler

import (
	"fmt"
	"megome/internal/config"
	"megome/internal/domain/refreshtoken"
	"megome/internal/pkg/httputil"
	"net/http"

	"github.com/gorilla/mux"
)

type RefreshTokenHandler struct {
	refreshStore *refreshtoken.Repository
}

func NewRefreshTokenHandler(refreshStore *refreshtoken.Repository) *RefreshTokenHandler {
	return &RefreshTokenHandler{refreshStore: refreshStore}
}

func (h *RefreshTokenHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/auth/refresh", h.handleRefresh).Methods("GET")
}

func (h *RefreshTokenHandler) handleRefresh(w http.ResponseWriter, r *http.Request) {
	refreshToken := httputil.GetTokenFromRequest(r)
	if refreshToken == "" {
		permissionDenied(w, "invalid token")
		return
	}

	newRefreshToken, newAccessToken, err := h.refreshStore.RefreshRotation(refreshToken)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid refresh token"))
		return
	}

	resp := refreshtoken.NewAuthResponse(
		"Token refreshed!",
		newAccessToken,
		newRefreshToken,
		config.Envs.JWTExpirationInSeconds,
	)
	httputil.WriteJSON(w, http.StatusOK, resp)
}
