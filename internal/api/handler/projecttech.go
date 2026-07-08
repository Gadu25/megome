package handler

import (
	"fmt"
	"megome/internal/domain/project"
	"megome/internal/domain/user"
	"megome/internal/middleware"
	"megome/internal/pkg/httputil"
	"megome/internal/pkg/validator"
	"net/http"
	"strconv"

	playvalidator "github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
)

type ProjectTechHandler struct {
	projectStore *project.Repository
	userStore    *user.Repository
}

type ProjectTechResponses struct {
	Message string              `json:"message"`
	Data    []project.ProjectTech `json:"data"`
}

type ValidationErrorResponse struct {
	Field string `json:"field"`
	Error string `json:"error"`
}

func NewProjectTechHandler(projectStore *project.Repository, userStore *user.Repository) *ProjectTechHandler {
	return &ProjectTechHandler{projectStore: projectStore, userStore: userStore}
}

func (h *ProjectTechHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/projectTech", middleware.WithJWTAuth(h.handleCreateProjectTech, h.userStore)).Methods("POST")
	router.HandleFunc("/projectTech/{id}/batch", middleware.WithJWTAuth(h.handleBatchCreateProjectTech, h.userStore)).Methods("POST")
	router.HandleFunc("/projectTech/{id}", middleware.WithJWTAuth(h.handleDeleteProjectTech, h.userStore)).Methods("DELETE")
}

func (h *ProjectTechHandler) handleCreateProjectTech(w http.ResponseWriter, r *http.Request) {
	projectID, err := strconv.Atoi(r.FormValue("projectId"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid projectId"))
		return
	}

	techID, err := strconv.Atoi(r.FormValue("techId"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid techId"))
		return
	}

	payload := project.ProjectTechPayload{
		ProjectID: projectID,
		TechID:    techID,
	}

	if err := httputil.ParseJSON(r, &payload); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err)
		return
	}

	if err := validator.Validate.Struct(payload); err != nil {
		errors := err.(playvalidator.ValidationErrors)
		httputil.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid payload %v", errors))
		return
	}

	err = h.projectStore.CreateProjectTech(project.ProjectTech{
		ProjectID: payload.ProjectID,
		TechID:    payload.TechID,
	})
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, map[string]string{
		"message": "Project tech successfully created",
	})
}

func (h *ProjectTechHandler) handleBatchCreateProjectTech(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.GetRequestId(r)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	var payload project.BatchProjectTechPayload

	if err := httputil.ParseJSON(r, &payload); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err)
		return
	}

	if err := validator.Validate.Struct(payload); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err)
		return
	}

	if err := h.projectStore.CreateProjectTechBatch(id, payload.TechIDs); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	httputil.WriteJSON(w, http.StatusOK, map[string]string{
		"message": "Technologies successfully linked to project",
	})
}

func (h *ProjectTechHandler) handleDeleteProjectTech(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.GetRequestId(r)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err)
		return
	}
	err = h.projectStore.DeleteProjectTech(id)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, map[string]string{
		"message": "Project tech is successfully deleted",
	})
}

func FormatValidationErrors(err error) []ValidationErrorResponse {
	var errors []ValidationErrorResponse

	validationErrors, ok := err.(playvalidator.ValidationErrors)
	if !ok {
		return []ValidationErrorResponse{
			{
				Error: "Invalid request payload",
			},
		}
	}

	for _, fieldErr := range validationErrors {
		switch fieldErr.Field() {

		case "TechIDs":
			switch fieldErr.Tag() {

			case "min":
				errors = append(errors, ValidationErrorResponse{
					Field: "techIds",
					Error: "At least one technology must be selected.",
				})

			default:
				errors = append(errors, ValidationErrorResponse{
					Field: "techIds",
					Error: "Invalid technologies payload.",
				})
			}

		default:
			errors = append(errors, ValidationErrorResponse{
				Field: fieldErr.Field(),
				Error: fmt.Sprintf(
					"Invalid value for %s.",
					fieldErr.Field(),
				),
			})
		}
	}

	return errors
}
