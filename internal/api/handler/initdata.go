package handler

import (
	"megome/internal/domain/profile"
	"megome/internal/domain/user"
	"megome/internal/middleware"
	"megome/internal/pkg/httputil"
	"net/http"

	"github.com/gorilla/mux"
)

type Onboarding struct {
	completed   bool
	currentStep string
}

type InitData struct {
	Profile *profile.Profile `json:"profile"`
}

type InitDataHandler struct {
	profileStore *profile.Repository
	userStore    *user.Repository
}

type InitResponse struct {
	Success bool `json:"success"`
	InitData
}

func NewInitDataHandler(profileStore *profile.Repository, userStore *user.Repository) *InitDataHandler {
	return &InitDataHandler{profileStore: profileStore, userStore: userStore}
}

func (h *InitDataHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/init", middleware.WithJWTAuth(h.handleInit, h.userStore)).Methods("GET")
}

func (h *InitDataHandler) handleInit(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r.Context())
	p, _ := h.profileStore.GetProfile(userID)

	init := InitData{
		Profile: p,
	}

	initResp := InitResponse{
		Success:  true,
		InitData: init,
	}

	httputil.WriteJSON(w, http.StatusOK, initResp)
}
