package init

import (
	"fmt"
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
	Profile *types.Profile
}

type Handler struct {
	profileStore types.ProfileStore
	userStore    types.UserStore
}

type InitResponse struct {
	Status bool
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
	profile, err := h.profileStore.GetProfile(userID)

	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("Error getting user: %w", err))
		return
	}

	init := InitData{
		Profile: profile,
	}

	initResp := InitResponse{
		Status:   true,
		InitData: init,
	}

	utils.WriteJSON(w, http.StatusOK, initResp)
}
