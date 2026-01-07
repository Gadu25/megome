package profile

import (
	"fmt"
	"megome/internal/services/auth"
	"megome/internal/services/types"
	"megome/internal/services/utils"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
)

type Handler struct {
	profileStore types.ProfileStore
	userStore    types.UserStore
}
type ProfileResponse struct {
	Message string         `json:"message"`
	Data    *types.Profile `json:"data"`
}

func NewHandler(profileStore types.ProfileStore, userStore types.UserStore) *Handler {
	return &Handler{profileStore: profileStore, userStore: userStore}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/profile", auth.WithJWTAuth(h.handleViewProfiles, h.userStore)).Methods("GET")
	router.HandleFunc("/profile", auth.WithJWTAuth(h.handleUpdateProfile, h.userStore)).Methods("POST")
}

func (h *Handler) handleViewProfiles(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserIDFromContext(r.Context())
	profile, err := h.profileStore.GetProfile(userID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}
	resp := ProfileResponse{
		Message: "Profile fetched successfully",
		Data:    profile,
	}
	utils.WriteJSON(w, http.StatusOK, resp)
}

func (h *Handler) handleUpdateProfile(w http.ResponseWriter, r *http.Request) {
	// get JSON payload
	var payload types.MakeProfilePayload
	if err := utils.ParseJSON(r, &payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	// validate the payload
	if err := utils.Validate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid payload %v", errors))
		return
	}

	// create or update Profile
	userID := auth.GetUserIDFromContext(r.Context())
	err := h.profileStore.MakeProfile(types.Profile{
		UserID:       userID,
		Bio:          payload.Bio,
		Phone:        payload.Phone,
		Website:      payload.Website,
		Location:     payload.Location,
		ProfileImage: payload.ProfileImage,
	})
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}
	utils.WriteJSON(w, http.StatusOK, map[string]string{"message": "Profile is successfully updated"})
}
