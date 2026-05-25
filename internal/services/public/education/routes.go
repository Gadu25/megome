package publiceducation

import (
	"megome/internal/platform/http/middleware"
	"megome/internal/services/auth"
	"megome/internal/services/types"
	"megome/internal/services/utils"
	"net/http"

	"github.com/gorilla/mux"
)

type Handler struct {
	educationStore types.EducationStore
	patStore       types.PersonalAccessTokenStore
	apiLogStore    types.APIUsageLogStore
}

type PublicResponse struct {
	Message   string            `json:"message"`
	Education []types.Education `json:"educations"`
}

func NewHandler(educationStore types.EducationStore, patStore types.PersonalAccessTokenStore, apiLogStore types.APIUsageLogStore) *Handler {
	return &Handler{
		educationStore: educationStore,
		patStore:       patStore,
		apiLogStore:    apiLogStore,
	}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/education",
		auth.WithPATAuth(
			middleware.WithAPILogging(h.handleGetPublicEducation, h.apiLogStore),
			h.patStore,
		),
	).Methods("GET")
}

func (h *Handler) handleGetPublicEducation(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetPATUserIDFromContext(r.Context())

	educations, err := h.educationStore.GetPublicEducations(userID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	resp := PublicResponse{
		Message:   "education successfully fetched",
		Education: educations,
	}

	utils.WriteJSON(w, http.StatusOK, resp)
}
