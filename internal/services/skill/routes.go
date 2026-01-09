package skill

import (
	"megome/internal/services/auth"
	"megome/internal/services/types"
	"net/http"

	"github.com/gorilla/mux"
)

type Handler struct {
	skillStore types.SkillStore
	userStore  types.UserStore
}

func NewHandler(skillStore types.SkillStore, userStore types.UserStore) *Handler {
	return &Handler{skillStore: skillStore, userStore: userStore}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/skill", auth.WithJWTAuth(h.handleViewSkills, h.userStore)).Methods("GET")
	router.HandleFunc("/skill", auth.WithJWTAuth(h.handleCreateSkill, h.userStore)).Methods("POST")
	router.HandleFunc("/skill/{id}", auth.WithJWTAuth(h.handleUpdateSkill, h.userStore)).Methods("PUT")
	router.HandleFunc("/skill/{id}", auth.WithJWTAuth(h.handleDeleteSkill, h.userStore)).Methods("DELETE")
}

func (h *Handler) handleViewSkills(w http.ResponseWriter, r *http.Request) {

}

func (h *Handler) handleCreateSkill(w http.ResponseWriter, r *http.Request) {

}

func (h *Handler) handleUpdateSkill(w http.ResponseWriter, r *http.Request) {

}

func (h *Handler) handleDeleteSkill(w http.ResponseWriter, r *http.Request) {

}
