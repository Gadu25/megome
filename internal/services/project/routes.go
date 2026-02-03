package project

import (
	"fmt"
	"megome/internal/services/auth"
	"megome/internal/services/types"
	"megome/internal/services/utils"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
)

type Handler struct {
	projectStore types.ProjectStore
	userStore    types.UserStore
}

type ProjectResponse struct {
	Message string          `json:"message"`
	Data    []types.Project `json:"project"`
}

func NewHandler(projectStore types.ProjectStore, userStore types.UserStore) *Handler {
	return &Handler{projectStore: projectStore, userStore: userStore}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/project", auth.WithJWTAuth(h.handleViewProject, h.userStore)).Methods("GET")
	router.HandleFunc("/project", auth.WithJWTAuth(h.handleCreateProject, h.userStore)).Methods("POST")
	router.HandleFunc("/project/{id}", auth.WithJWTAuth(h.handleUpdateProject, h.userStore)).Methods("PUT")
	router.HandleFunc("/project/{id}", auth.WithJWTAuth(h.handleDeleteProject, h.userStore)).Methods("DELETE")
}

func (h *Handler) handleViewProject(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserIDFromContext(r.Context())
	projects, err := h.projectStore.GetProjects(userID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}
	resp := ProjectResponse{
		Message: "Project fetched successfully",
		Data:    projects,
	}
	utils.WriteJSON(w, http.StatusOK, resp)
}

func (h *Handler) handleCreateProject(w http.ResponseWriter, r *http.Request) {
	// get JSON payload
	var payload types.ProjectPayload
	if err := utils.ParseJSON(r, &payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	// validate the payload
	if err := utils.Validate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid payload %v", errors))
		return
	}

	// create project
	userID := auth.GetUserIDFromContext(r.Context())
	err := h.projectStore.CreateProject(types.Project{
		UserID:      userID,
		Title:       payload.Title,
		Description: payload.Description,
		Link:        payload.Link,
		GithubLink:  payload.GithubLink,
	})
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}
	utils.WriteJSON(w, http.StatusOK, map[string]string{
		"message": "Project is successfully created",
	})
}

func (h *Handler) handleUpdateProject(w http.ResponseWriter, r *http.Request) {
	// get JSON payload
	var payload types.ProjectPayload
	if err := utils.ParseJSON(r, &payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	// validate the payload
	if err := utils.Validate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid payload %v", errors))
		return
	}

	// update project
	id, err := utils.GetRequestId(r)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}
	err = h.projectStore.UpdateProject(id, types.Project{
		Title:       payload.Title,
		Description: payload.Description,
		Link:        payload.Link,
		GithubLink:  payload.GithubLink,
	})
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}
	utils.WriteJSON(w, http.StatusOK, map[string]string{
		"message": "Project is successfully updated",
	})
}

func (h *Handler) handleDeleteProject(w http.ResponseWriter, r *http.Request) {
	id, err := utils.GetRequestId(r)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}
	err = h.projectStore.DeleteProject(id)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}
	utils.WriteJSON(w, http.StatusOK, map[string]string{
		"message": "Project is successfully deleted",
	})
}
