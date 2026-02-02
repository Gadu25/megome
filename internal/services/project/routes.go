package project

import (
	"megome/internal/services/auth"
	"megome/internal/services/types"
	"net/http"

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

}

func (h *Handler) handleCreateProject(w http.ResponseWriter, r *http.Request) {

}

func (h *Handler) handleUpdateProject(w http.ResponseWriter, r *http.Request) {

}

func (h *Handler) handleDeleteProject(w http.ResponseWriter, r *http.Request) {

}
