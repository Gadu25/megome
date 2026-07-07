package public

import (
	"megome/internal/domain/apilog"
	"megome/internal/domain/personalaccesstoken"
	"megome/internal/domain/skill"
	"megome/internal/middleware"
	"megome/internal/pkg/httputil"
	"net/http"

	"github.com/gorilla/mux"
)

type SkillHandler struct {
	skillStore  *skill.Repository
	patStore    *personalaccesstoken.Repository
	apiLogStore *apilog.Repository
}

type SkillPublicResponse struct {
	Message string        `json:"message"`
	Skills  []skill.Skill `json:"skills"`
}

func NewSkillHandler(skillStore *skill.Repository, patStore *personalaccesstoken.Repository, apiLogStore *apilog.Repository) *SkillHandler {
	return &SkillHandler{
		skillStore:  skillStore,
		patStore:    patStore,
		apiLogStore: apiLogStore,
	}
}

func (h *SkillHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/skill",
		middleware.WithPATAuth(
			middleware.WithAPILogging(h.handleGetPublicSkill, h.apiLogStore),
			h.patStore,
		),
	).Methods("GET")
}

func (h *SkillHandler) handleGetPublicSkill(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetPATUserIDFromContext(r.Context())

	skills, err := h.skillStore.GetPublicSkills(userID)

	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	resp := SkillPublicResponse{
		Message: "skills successfully fetched",
		Skills:  skills,
	}

	httputil.WriteJSON(w, http.StatusOK, resp)
}
