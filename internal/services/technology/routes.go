package technology

import (
	"megome/internal/services/auth"
	"megome/internal/services/types"
	"net/http"

	"github.com/gorilla/mux"
)

type Handler struct {
	technologyStore types.TechnologyStore
	userStore       types.UserStore
}

type TechnologyResponses struct {
	Message string             `json:"message"`
	Data    []types.Technology `json:"data"`
}

func NewHandler(technologyStore types.TechnologyStore, userStore types.UserStore) *Handler {
	return &Handler{technologyStore: technologyStore, userStore: userStore}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/technology", auth.WithJWTAuth(h.handleViewTechnology, h.userStore)).Methods("GET")
	router.HandleFunc("/technology", auth.WithJWTAuth(h.handleCreateTechnology, h.userStore)).Methods("POST")
	router.HandleFunc("/technology/{id}", auth.WithJWTAuth(h.handleUpdateTechnology, h.userStore)).Methods("PUT")
	router.HandleFunc("/technology/{id}", auth.WithJWTAuth(h.handleDeleteTechnology, h.userStore)).Methods("DELETE")
}

func (h *Handler) handleViewTechnology(w http.ResponseWriter, r *http.Request) {

}
func (h *Handler) handleCreateTechnology(w http.ResponseWriter, r *http.Request) {

}
func (h *Handler) handleUpdateTechnology(w http.ResponseWriter, r *http.Request) {

}
func (h *Handler) handleDeleteTechnology(w http.ResponseWriter, r *http.Request) {

}
