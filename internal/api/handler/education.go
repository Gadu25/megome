package handler

import (
	"fmt"
	"megome/internal/domain/education"
	"megome/internal/domain/user"
	"megome/internal/middleware"
	"megome/internal/pkg/httputil"
	"megome/internal/pkg/validator"
	"net/http"

	playvalidator "github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
)

type EducationHandler struct {
	educationStore *education.Repository
	userStore      *user.Repository
}

type EducationResponse struct {
	Message   string              `json:"message"`
	Education []education.Education `json:"educations"`
}

type SingleEducResponse struct {
	Message   string             `json:"message"`
	Education education.Education `json:"education"`
}

func NewEducationHandler(educationStore *education.Repository, userStore *user.Repository) *EducationHandler {
	return &EducationHandler{educationStore: educationStore, userStore: userStore}
}

func (h *EducationHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/education", middleware.WithJWTAuth(h.handleViewEducation, h.userStore)).Methods("GET")
	router.HandleFunc("/education", middleware.WithJWTAuth(h.handleCreateEducation, h.userStore)).Methods("POST")
	router.HandleFunc("/education/{id}", middleware.WithJWTAuth(h.handleEditEducation, h.userStore)).Methods("PUT")
	router.HandleFunc("/education/{id}", middleware.WithJWTAuth(h.handleDeleteEducation, h.userStore)).Methods("DELETE")
}

func (h *EducationHandler) handleViewEducation(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r.Context())
	educations, err := h.educationStore.GetEducations(userID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err)
		return
	}
	resp := EducationResponse{
		Message:   "Education fetched successfully",
		Education: educations,
	}
	httputil.WriteJSON(w, http.StatusOK, resp)
}

func (h *EducationHandler) handleCreateEducation(w http.ResponseWriter, r *http.Request) {
	var payload education.EducationPayload
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
	educ, err := h.educationStore.CreateEducation(education.Education{
		UserID:       userID,
		School:       payload.School,
		Description:  payload.Description,
		Degree:       payload.Degree,
		FieldOfStudy: payload.FieldOfStudy,
		StartDate:    payload.StartDate,
		EndDate:      payload.EndDate,
		IsPresent:    payload.IsPresent,
	})
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	resp := SingleEducResponse{
		Message:   "Education added successfully",
		Education: educ,
	}
	httputil.WriteJSON(w, http.StatusOK, resp)
}

func (h *EducationHandler) handleEditEducation(w http.ResponseWriter, r *http.Request) {
	var payload education.EducationPayload
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

	educ, err := h.educationStore.UpdateEducation(id, education.Education{
		School:       payload.School,
		Description:  payload.Description,
		Degree:       payload.Degree,
		FieldOfStudy: payload.FieldOfStudy,
		StartDate:    payload.StartDate,
		EndDate:      payload.EndDate,
		IsPresent:    payload.IsPresent,
	})
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	resp := SingleEducResponse{
		Message:   "Education updated successfully",
		Education: educ,
	}
	httputil.WriteJSON(w, http.StatusOK, resp)
}

func (h *EducationHandler) handleDeleteEducation(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.GetRequestId(r)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err)
		return
	}

	educ, err := h.educationStore.DeleteEducation(id)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	resp := SingleEducResponse{
		Message:   "Education deleted successfully",
		Education: educ,
	}
	httputil.WriteJSON(w, http.StatusOK, resp)
}
