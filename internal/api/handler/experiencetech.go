package handler

import (
	"fmt"
	"megome/internal/domain/experience"
	"megome/internal/domain/user"
	"megome/internal/middleware"
	"megome/internal/pkg/httputil"
	"megome/internal/pkg/validator"
	"net/http"
	"strconv"

	playvalidator "github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
)

type ExperienceTechHandler struct {
	experienceStore *experience.Repository
	userStore       *user.Repository
}

type ExperienceTechResponse struct {
	Message string `json:"message"`
}

func NewExperienceTechHandler(experienceStore *experience.Repository, userStore *user.Repository) *ExperienceTechHandler {
	return &ExperienceTechHandler{experienceStore: experienceStore, userStore: userStore}
}

func (h *ExperienceTechHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/experienceTech", middleware.WithJWTAuth(h.handleCreateExperienceTech, h.userStore)).Methods("POST")
	router.HandleFunc("/experienceTech/{id}/batch", middleware.WithJWTAuth(h.handleBatchCreateExperienceTech, h.userStore)).Methods("POST")
	router.HandleFunc("/experienceTech/{id}", middleware.WithJWTAuth(h.handleDeleteExperienceTech, h.userStore)).Methods("DELETE")
}

func (h *ExperienceTechHandler) handleCreateExperienceTech(w http.ResponseWriter, r *http.Request) {
	experienceID, err := strconv.Atoi(r.FormValue("experienceId"))
	if err != nil {
		http.Error(w, "invalid experienceId", http.StatusBadRequest)
		return
	}

	techID, err := strconv.Atoi(r.FormValue("techId"))
	if err != nil {
		http.Error(w, "invalid techId", http.StatusBadRequest)
		return
	}

	payload := experience.ExperienceTechPayload{
		ExperienceID: experienceID,
		TechID:       techID,
	}

	if err := validator.Validate.Struct(payload); err != nil {
		errors := err.(playvalidator.ValidationErrors)
		httputil.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid payload %v", errors))
		return
	}

	err = h.experienceStore.CreateExperienceTech(experience.ExperienceTech{
		ExperienceID: payload.ExperienceID,
		TechID:       payload.TechID,
	})
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	httputil.WriteJSON(w, http.StatusOK, ExperienceTechResponse{
		Message: "Technology successfully linked to experience",
	})
}

func (h *ExperienceTechHandler) handleBatchCreateExperienceTech(w http.ResponseWriter, r *http.Request) {
	expId, err := httputil.GetRequestId(r)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	var payload experience.BatchExperienceTechPayload

	if err := httputil.ParseJSON(r, &payload); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err)
		return
	}

	if err := validator.Validate.Struct(payload); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err)
		return
	}

	if err := h.experienceStore.CreateExperienceTechBatch(expId, payload.TechIDs); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	httputil.WriteJSON(w, http.StatusOK, ExperienceTechResponse{
		Message: "Technologies successfully linked to experience",
	})
}

func (h *ExperienceTechHandler) handleDeleteExperienceTech(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.GetRequestId(r)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	err = h.experienceStore.DeleteExperienceTech(id)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	httputil.WriteJSON(w, http.StatusOK, ExperienceTechResponse{
		Message: "Experience tech successfully deleted",
	})
}
