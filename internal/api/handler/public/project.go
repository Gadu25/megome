package public

import (
	"megome/internal/domain/apilog"
	"megome/internal/domain/personalaccesstoken"
	"megome/internal/domain/project"
	"megome/internal/middleware"
	"megome/internal/pkg/httputil"
	"net/http"

	"github.com/gorilla/mux"
)

type ProjectHandler struct {
	projectStore *project.Repository
	patStore     *personalaccesstoken.Repository
	apiLogStore  *apilog.Repository
}

type ProjectPublicResponse struct {
	Message  string                `json:"message"`
	Projects []project.ProjectFull `json:"projects"`
}

type SingleProjPublicResponse struct {
	Message string               `json:"message"`
	Project project.ProjectFull  `json:"project"`
}

func NewProjectHandler(projectStore *project.Repository, patStore *personalaccesstoken.Repository, apiLogStore *apilog.Repository) *ProjectHandler {
	return &ProjectHandler{
		projectStore: projectStore,
		patStore:     patStore,
		apiLogStore:  apiLogStore,
	}
}

func (h *ProjectHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/project",
		middleware.WithPATAuth(
			middleware.WithAPILogging(h.handleGetPublicProjects, h.apiLogStore),
			h.patStore,
		),
	).Methods("GET")
	router.HandleFunc("/project/{id}",
		middleware.WithPATAuth(
			middleware.WithAPILogging(h.handleGetPublicProject, h.apiLogStore),
			h.patStore,
		),
	).Methods("GET")
}

func (h *ProjectHandler) handleGetPublicProjects(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetPATUserIDFromContext(r.Context())

	projects, err := h.projectStore.GetProjectsFull(userID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	resp := ProjectPublicResponse{
		Message:  "projects successfully fetched",
		Projects: projects,
	}

	httputil.WriteJSON(w, http.StatusOK, resp)
}

func (h *ProjectHandler) handleGetPublicProject(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.GetRequestId(r)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	p, err := h.projectStore.GetProjectById(id)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err)
		return
	}
	resp := SingleProjPublicResponse{
		Message: "Project fetched successfully",
		Project: p,
	}
	httputil.WriteJSON(w, http.StatusOK, resp)
}
