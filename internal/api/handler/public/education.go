package public

import (
	"megome/internal/domain/apilog"
	"megome/internal/domain/education"
	"megome/internal/domain/personalaccesstoken"
	"megome/internal/middleware"
	"megome/internal/pkg/httputil"
	"net/http"

	"github.com/gorilla/mux"
)

type EducationHandler struct {
	educationStore *education.Repository
	patStore       *personalaccesstoken.Repository
	apiLogStore    *apilog.Repository
}

type EducationPublicResponse struct {
	Message   string               `json:"message"`
	Education []education.Education `json:"educations"`
}

func NewEducationHandler(educationStore *education.Repository, patStore *personalaccesstoken.Repository, apiLogStore *apilog.Repository) *EducationHandler {
	return &EducationHandler{
		educationStore: educationStore,
		patStore:       patStore,
		apiLogStore:    apiLogStore,
	}
}

func (h *EducationHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/education",
		middleware.WithPATAuth(
			middleware.WithAPILogging(h.handleGetPublicEducation, h.apiLogStore),
			h.patStore,
		),
	).Methods("GET")
}

func (h *EducationHandler) handleGetPublicEducation(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetPATUserIDFromContext(r.Context())

	educations, err := h.educationStore.GetEducations(userID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	resp := EducationPublicResponse{
		Message:   "education successfully fetched",
		Education: educations,
	}

	httputil.WriteJSON(w, http.StatusOK, resp)
}
