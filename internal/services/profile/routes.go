package profile

import (
	"megome/internal/services/auth"
	"megome/internal/services/types"
	"megome/internal/services/utils"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type Handler struct {
	profileStore types.ProfileStore
	userStore    types.UserStore
}

func NewHandler(profileStore types.ProfileStore, userStore types.UserStore) *Handler {
	return &Handler{profileStore: profileStore, userStore: userStore}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/profiles", auth.WithJWTAuth(h.handleViewProfiles, h.userStore)).Methods("GET")
	// router.HandleFunc("/profiles", h.handleViewProfiles).Methods("GET")
}

func (h *Handler) handleViewProfiles(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserIDFromContext(r.Context())
	utils.WriteJSON(w, http.StatusOK, map[string]string{"userId": strconv.Itoa(userID)})
	// utils.WriteJSON(w, http.StatusOK, map[string]string{"userId": "testing"})
}
