package handler

import (
	"fmt"
	"megome/internal/domain/project"
	"megome/internal/domain/user"
	"megome/internal/middleware"
	"megome/internal/pkg/httputil"
	"megome/internal/pkg/validator"
	"net/http"

	playvalidator "github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
)

type ProjectHandler struct {
	projectStore *project.Repository
	userStore    *user.Repository
}

type ProjectResponse struct {
	Message  string           `json:"message"`
	Projects []project.Project `json:"projects"`
}

type FullProjectResponse struct {
	Message  string               `json:"message"`
	Projects []project.ProjectFull `json:"projects"`
}

type SingleProjResponse struct {
	Message string             `json:"message"`
	Project project.ProjectFull `json:"project"`
}

func NewProjectHandler(projectStore *project.Repository, userStore *user.Repository) *ProjectHandler {
	return &ProjectHandler{projectStore: projectStore, userStore: userStore}
}

func (h *ProjectHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/project", middleware.WithJWTAuth(h.handleViewProjects, h.userStore)).Methods("GET")
	router.HandleFunc("/project/reorder", middleware.WithJWTAuth(h.handleReorderProjects, h.userStore)).Methods("POST")
	router.HandleFunc("/project/{id}", middleware.WithJWTAuth(h.handleViewProject, h.userStore)).Methods("GET")
	router.HandleFunc("/project", middleware.WithJWTAuth(h.handleCreateProject, h.userStore)).Methods("POST")
	router.HandleFunc("/project/{id}", middleware.WithJWTAuth(h.handleUpdateProject, h.userStore)).Methods("PUT")
	router.HandleFunc("/project/{id}", middleware.WithJWTAuth(h.handleDeleteProject, h.userStore)).Methods("DELETE")
}

func (h *ProjectHandler) handleReorderProjects(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Items []project.ReorderItem `json:"items"`
	}
	if err := httputil.ParseJSON(r, &payload); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err)
		return
	}

	if len(payload.Items) == 0 {
		httputil.WriteError(w, http.StatusBadRequest, fmt.Errorf("items list cannot be empty"))
		return
	}

	userID := middleware.GetUserIDFromContext(r.Context())

	if err := h.projectStore.ReorderProjects(userID, payload.Items); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	httputil.WriteJSON(w, http.StatusOK, map[string]string{"message": "Projects reordered successfully"})
}

func (h *ProjectHandler) handleViewProject(w http.ResponseWriter, r *http.Request) {
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
	resp := SingleProjResponse{
		Message: "Project fetched successfully",
		Project: p,
	}
	httputil.WriteJSON(w, http.StatusOK, resp)
}

func (h *ProjectHandler) handleViewProjects(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r.Context())
	projects, err := h.projectStore.GetProjectsFull(userID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err)
		return
	}
	resp := FullProjectResponse{
		Message:  "Project fetched successfully",
		Projects: projects,
	}
	httputil.WriteJSON(w, http.StatusOK, resp)
}

func (h *ProjectHandler) handleCreateProject(w http.ResponseWriter, r *http.Request) {
	var payload project.ProjectPayload
	if err := httputil.ParseJSON(r, &payload); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err)
		return
	}

	if err := validator.Validate.Struct(payload); err != nil {
		errors := err.(playvalidator.ValidationErrors)
		httputil.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid payload %v", errors))
		return
	}

	userID := middleware.GetUserIDFromContext(r.Context())
	p, err := h.projectStore.CreateProject(project.Project{
		UserID:      userID,
		Title:       payload.Title,
		Status:      payload.Status,
		Description: payload.Description,
		Link:        payload.Link,
		GithubLink:  payload.GithubLink,
	})

	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	resp := SingleProjResponse{
		Message: "Project created successfully",
		Project: p,
	}
	httputil.WriteJSON(w, http.StatusOK, resp)
}

func (h *ProjectHandler) handleUpdateProject(w http.ResponseWriter, r *http.Request) {
	var payload project.ProjectPayload
	if err := httputil.ParseJSON(r, &payload); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err)
		return
	}

	if err := validator.Validate.Struct(payload); err != nil {
		errors := err.(playvalidator.ValidationErrors)
		httputil.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid payload %v", errors))
		return
	}

	id, err := httputil.GetRequestId(r)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err)
		return
	}
	p, err := h.projectStore.UpdateProject(id, project.Project{
		Title:       payload.Title,
		Status:      payload.Status,
		Description: payload.Description,
		Link:        payload.Link,
		GithubLink:  payload.GithubLink,
		IsDraft:     payload.IsDraft,
	})
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	resp := SingleProjResponse{
		Message: "Project updated successfully",
		Project: p,
	}
	httputil.WriteJSON(w, http.StatusOK, resp)
}

func (h *ProjectHandler) handleDeleteProject(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.GetRequestId(r)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err)
		return
	}
	p, err := h.projectStore.DeleteProject(id)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	resp := SingleProjResponse{
		Message: "Project deleted successfully",
		Project: p,
	}
	httputil.WriteJSON(w, http.StatusOK, resp)
}
