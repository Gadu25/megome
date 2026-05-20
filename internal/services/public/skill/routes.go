package publicskill

import (
	"megome/internal/services/auth"
	"megome/internal/services/middleware"
	"megome/internal/services/types"
	"megome/internal/services/utils"
	"net/http"

	"github.com/gorilla/mux"
)

type Handler struct {
	skillStore  types.SkillStore
	patStore    types.PersonalAccessTokenStore
	apiLogStore types.APIUsageLogStore
}

type PublicResponse struct {
	Message string        `json:"message"`
	Skills  []types.Skill `json:"skills"`
}

func NewHandler(skillStore types.SkillStore, patStore types.PersonalAccessTokenStore, apiLogStore types.APIUsageLogStore) *Handler {
	return &Handler{
		skillStore:  skillStore,
		patStore:    patStore,
		apiLogStore: apiLogStore,
	}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/skill",
		auth.WithPATAuth(
			middleware.WithAPILogging(h.handleGetPublicSkill, h.apiLogStore),
			h.patStore,
		),
	).Methods("GET")
}

func (h *Handler) handleGetPublicSkill(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetPATUserIDFromContext(r.Context())

	skills, err := h.skillStore.GetPublicSkills(userID)

	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	resp := PublicResponse{
		Message: "profile successfully fetched",
		Skills:  skills,
	}

	utils.WriteJSON(w, http.StatusOK, resp)
}
