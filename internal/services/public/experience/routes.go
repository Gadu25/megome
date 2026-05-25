package publicexperience

import (
	"megome/internal/platform/http/middleware"
	"megome/internal/services/auth"
	"megome/internal/services/types"
	"megome/internal/services/utils"
	"net/http"

	"github.com/gorilla/mux"
)

type Handler struct {
	experienceStore types.ExperienceStore
	patStore        types.PersonalAccessTokenStore
	apiLogStore     types.APIUsageLogStore
}

type PublicResponse struct {
	Message     string             `json:"message"`
	Experiences []types.Experience `json:"experiences"`
}

func NewHandler(experienceStore types.ExperienceStore, patStore types.PersonalAccessTokenStore, apiLogStore types.APIUsageLogStore) *Handler {
	return &Handler{
		experienceStore: experienceStore,
		patStore:        patStore,
		apiLogStore:     apiLogStore,
	}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/experience",
		auth.WithPATAuth(
			middleware.WithAPILogging(h.handleGetPublicEducation, h.apiLogStore),
			h.patStore,
		),
	).Methods("GET")
}

func (h *Handler) handleGetPublicEducation(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetPATUserIDFromContext(r.Context())

	experiences, err := h.experienceStore.GetPublicExperiences(userID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	resp := PublicResponse{
		Message:     "experiences successfully fetched",
		Experiences: experiences,
	}

	utils.WriteJSON(w, http.StatusOK, resp)
}
