package public

import (
	"megome/internal/domain/apilog"
	"megome/internal/domain/experience"
	"megome/internal/domain/personalaccesstoken"
	"megome/internal/middleware"
	"megome/internal/pkg/httputil"
	"net/http"

	"github.com/gorilla/mux"
)

type ExperienceHandler struct {
	experienceStore *experience.Repository
	patStore        *personalaccesstoken.Repository
	apiLogStore     *apilog.Repository
}

type ExperiencePublicResponse struct {
	Message     string                 `json:"message"`
	Experiences []experience.Experience `json:"experiences"`
}

func NewExperienceHandler(experienceStore *experience.Repository, patStore *personalaccesstoken.Repository, apiLogStore *apilog.Repository) *ExperienceHandler {
	return &ExperienceHandler{
		experienceStore: experienceStore,
		patStore:        patStore,
		apiLogStore:     apiLogStore,
	}
}

func (h *ExperienceHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/experience",
		middleware.WithPATAuth(
			middleware.WithAPILogging(h.handleGetPublicExperience, h.apiLogStore),
			h.patStore,
		),
	).Methods("GET")
}

func (h *ExperienceHandler) handleGetPublicExperience(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetPATUserIDFromContext(r.Context())

	experiences, err := h.experienceStore.GetPublicExperiences(userID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	resp := ExperiencePublicResponse{
		Message:     "experiences successfully fetched",
		Experiences: experiences,
	}

	httputil.WriteJSON(w, http.StatusOK, resp)
}
