package handler

import (
	"fmt"
	"megome/internal/domain/technology"
	"megome/internal/domain/user"
	"megome/internal/middleware"
	"megome/internal/pkg/httputil"
	"megome/internal/pkg/validator"
	"net/http"

	playvalidator "github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
)

type TechnologyHandler struct {
	technologyStore *technology.Repository
	userStore       *user.Repository
}

type TechnologyResponses struct {
	Message      string                `json:"message"`
	Technologies []technology.Technology `json:"technologies"`
}

func NewTechnologyHandler(technologyStore *technology.Repository, userStore *user.Repository) *TechnologyHandler {
	return &TechnologyHandler{technologyStore: technologyStore, userStore: userStore}
}

func (h *TechnologyHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/technology", middleware.WithJWTAuth(h.handleViewTechnology, h.userStore)).Methods("GET")
	router.HandleFunc("/technology", middleware.WithJWTAuth(h.handleCreateTechnology, h.userStore)).Methods("POST")
	router.HandleFunc("/technology/{id}", middleware.WithJWTAuth(h.handleUpdateTechnology, h.userStore)).Methods("PUT")
	router.HandleFunc("/technology/{id}", middleware.WithJWTAuth(h.handleDeleteTechnology, h.userStore)).Methods("DELETE")
}

func (h *TechnologyHandler) handleViewTechnology(w http.ResponseWriter, r *http.Request) {
	technologies, err := h.technologyStore.GetTechnologies()
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err)
		return
	}
	resp := TechnologyResponses{
		Message:      "Technology fetched successfully",
		Technologies: technologies,
	}
	httputil.WriteJSON(w, http.StatusOK, resp)
}

func (h *TechnologyHandler) handleCreateTechnology(w http.ResponseWriter, r *http.Request) {
	var payload technology.TechnologyPayload
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
	err := h.technologyStore.CreateTechnology(technology.Technology{
		CreatedByUserId: &userID,
		Name:            payload.Name,
		Category:        payload.Category,
	})
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, map[string]string{
		"message": "Technology is successfully created",
	})
}

func (h *TechnologyHandler) handleUpdateTechnology(w http.ResponseWriter, r *http.Request) {
	var payload technology.TechnologyPayload
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
	err = h.technologyStore.UpdateTechnology(id, technology.Technology{
		Name:     payload.Name,
		Category: payload.Category,
	})
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, map[string]string{"message": "Technology is successfully updated"})
}

func (h *TechnologyHandler) handleDeleteTechnology(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.GetRequestId(r)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err)
		return
	}
	err = h.technologyStore.DeleteTechnology(id)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, map[string]string{"message": "Technology is successfully deleted"})
}
