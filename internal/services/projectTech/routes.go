package projecttech

import (
	"fmt"
	"megome/internal/services/auth"
	"megome/internal/services/types"
	"megome/internal/services/utils"
	"net/http"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
)

type Handler struct {
	projectTechStore types.ProjectTechStore
	userStore        types.UserStore
}

type ProjectResponses struct {
	Message string              `json:"message"`
	Data    []types.ProjectTech `json:"data"`
}

func NewHandler(projectTechStore types.ProjectTechStore, userStore types.UserStore) *Handler {
	return &Handler{projectTechStore: projectTechStore, userStore: userStore}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/projectTech", auth.WithJWTAuth(h.handleCreateProjectTech, h.userStore)).Methods("POST")
	// in the future that we need updating
	// router.HandleFunc("/projectTech/{id}", auth.WithJWTAuth(h.handleUpdateProjectTech, h.userStore)).Methods("PUT")
	router.HandleFunc("/projectTech/{id}/batch", auth.WithJWTAuth(h.handleBatchCreateProjectTech, h.userStore)).Methods("POST")
	router.HandleFunc("/projectTech/{id}", auth.WithJWTAuth(h.handleDeleteProjectTech, h.userStore)).Methods("DELETE")
}

func (h *Handler) handleCreateProjectTech(w http.ResponseWriter, r *http.Request) {
	// get JSON payload
	projectID, err := strconv.Atoi(r.FormValue("projectId"))
	if err != nil {
		http.Error(w, "invalid projectId", http.StatusBadRequest)
		return
	}

	techID, err := strconv.Atoi(r.FormValue("techId"))
	if err != nil {
		http.Error(w, "invalid techId", http.StatusBadRequest)
		return
	}

	payload := types.ProjectTechPayload{
		ProjectID: projectID,
		TechID:    techID,
	}

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

	// create certification
	err = h.projectTechStore.CreateProjectTech(types.ProjectTech{
		ProjectID: payload.ProjectID,
		TechID:    payload.TechID,
	})
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}
	utils.WriteJSON(w, http.StatusOK, map[string]string{
		"message": "Certification is successfully created",
	})
}

func (h *Handler) handleBatchCreateProjectTech(w http.ResponseWriter, r *http.Request) {
	id, err := utils.GetRequestId(r)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	var payload types.BatchProjectTechPayload

	// decode JSON
	if err := utils.ParseJSON(r, &payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	// validate payload
	if err := utils.Validate.Struct(payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	if err := h.projectTechStore.CreateProjectTechBatch(id, payload.TechIDs); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]string{
		"message": "Technologies successfully linked to project",
	})
}

func (h *Handler) handleDeleteProjectTech(w http.ResponseWriter, r *http.Request) {
	id, err := utils.GetRequestId(r)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}
	err = h.projectTechStore.DelteProjectTech(id)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}
	utils.WriteJSON(w, http.StatusOK, map[string]string{
		"message": "Project tech is successfully deleted",
	})
}
