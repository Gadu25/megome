package publicprofile

import (
	"megome/internal/platform/http/middleware"
	"megome/internal/services/auth"
	"megome/internal/services/types"
	"megome/internal/services/utils"
	"net/http"

	"github.com/gorilla/mux"
)

type Handler struct {
	profileStore types.ProfileStore
	patStore     types.PersonalAccessTokenStore
	apiLogStore  types.APIUsageLogStore
}

type PublicResponse struct {
	Message string         `json:"message"`
	Data    *types.Profile `json:"data"`
}

func NewHandler(profileStore types.ProfileStore, patStore types.PersonalAccessTokenStore, apiLogStore types.APIUsageLogStore) *Handler {
	return &Handler{
		profileStore: profileStore,
		patStore:     patStore,
		apiLogStore:  apiLogStore,
	}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/profile",
		auth.WithPATAuth(
			middleware.WithAPILogging(h.handleGetPublicProfile, h.apiLogStore),
			h.patStore,
		),
	).Methods("GET")
}

func (h *Handler) handleGetPublicProfile(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetPATUserIDFromContext(r.Context())

	profile, err := h.profileStore.GetPublicProfile(userID)

	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	resp := PublicResponse{
		Message: "profile successfully fetched",
		Data:    profile,
	}

	utils.WriteJSON(w, http.StatusOK, resp)
}
