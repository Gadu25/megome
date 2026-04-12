package initData

import (
	"megome/internal/services/auth"
	"megome/internal/services/types"
	"megome/internal/services/utils"
	"net/http"

	"github.com/gorilla/mux"
)

type Onboarding struct {
	completed   bool
	currentStep string
}

type InitData struct {
	Profile *types.Profile `json:"profile"`
}

type Handler struct {
	profileStore types.ProfileStore
	userStore    types.UserStore
}

type InitResponse struct {
	success bool
	InitData
}

func NewHandler(profileStore types.ProfileStore, userStore types.UserStore) *Handler {
	return &Handler{profileStore: profileStore, userStore: userStore}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/init", auth.WithJWTAuth(h.handleInit, h.userStore)).Methods("GET")
}

func (h *Handler) handleInit(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserIDFromContext(r.Context())
	profile, _ := h.profileStore.GetProfile(userID)

	init := InitData{
		Profile: profile,
	}

	initResp := InitResponse{
		success:  true,
		InitData: init,
	}

	utils.WriteJSON(w, http.StatusOK, initResp)
}
