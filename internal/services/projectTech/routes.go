package projecttech

import (
	"megome/internal/services/auth"
	"megome/internal/services/types"
	"net/http"

	"github.com/gorilla/mux"
)

type Handler struct {
	projectTechStore types.ProjectTechStore
	userStore        types.UserStore
}

func NewHandler(projectTechStore types.ProjectTechStore, userStore types.UserStore) *Handler {
	return &Handler{projectTechStore: projectTechStore, userStore: userStore}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/projectTech", auth.WithJWTAuth(h.handleViewProjectTech, h.userStore)).Methods("GET")
	router.HandleFunc("/projectTech", auth.WithJWTAuth(h.handleCreateProjectTech, h.userStore)).Methods("POST")
	router.HandleFunc("/projectTech/{id}", auth.WithJWTAuth(h.handleUpdateProjectTech, h.userStore)).Methods("PUT")
	router.HandleFunc("/projectTech/{id}", auth.WithJWTAuth(h.handleDeleteProjectTech, h.userStore)).Methods("DELETE")
}

func (h *Handler) handleViewProjectTech(w http.ResponseWriter, r *http.Request) {

}
func (h *Handler) handleCreateProjectTech(w http.ResponseWriter, r *http.Request) {

}
func (h *Handler) handleUpdateProjectTech(w http.ResponseWriter, r *http.Request) {

}
func (h *Handler) handleDeleteProjectTech(w http.ResponseWriter, r *http.Request) {

}
