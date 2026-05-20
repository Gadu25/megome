package publicproject

import (
	"megome/internal/services/auth"
	"megome/internal/services/middleware"
	"megome/internal/services/types"
	"megome/internal/services/utils"
	"net/http"

	"github.com/gorilla/mux"
)

type Handler struct {
	projectStore types.ProjectStore
	patStore     types.PersonalAccessTokenStore
	apiLogStore  types.APIUsageLogStore
}

type PublicResponse struct {
	Message  string              `json:"message"`
	Projects []types.ProjectFull `json:"projects"`
}

func NewHandler(projectStore types.ProjectStore, patStore types.PersonalAccessTokenStore, apiLogStore types.APIUsageLogStore) *Handler {
	return &Handler{
		projectStore: projectStore,
		patStore:     patStore,
		apiLogStore:  apiLogStore,
	}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/project",
		auth.WithPATAuth(
			middleware.WithAPILogging(h.handleGetPublicEducation, h.apiLogStore),
			h.patStore,
		),
	).Methods("GET")
}

func (h *Handler) handleGetPublicEducation(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetPATUserIDFromContext(r.Context())

	projects, err := h.projectStore.GetPublicProjects(userID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	resp := PublicResponse{
		Message:  "projects successfully fetched",
		Projects: projects,
	}

	utils.WriteJSON(w, http.StatusOK, resp)
}
