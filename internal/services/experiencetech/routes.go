package experiencetech

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
	experienceTechStore types.ExperienceTechStore
	userStore           types.UserStore
}

type ExperienceTechResponse struct {
	Message string `json:"message"`
}

func NewHandler(experienceTechStore types.ExperienceTechStore, userStore types.UserStore) *Handler {
	return &Handler{experienceTechStore: experienceTechStore, userStore: userStore}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/experienceTech", auth.WithJWTAuth(h.handleCreateExperienceTech, h.userStore)).Methods("POST")
	router.HandleFunc("/experienceTech/{id}/batch", auth.WithJWTAuth(h.handleBatchCreateExperienceTech, h.userStore)).Methods("POST")
	router.HandleFunc("/experienceTech/{id}", auth.WithJWTAuth(h.handleDeleteExperienceTech, h.userStore)).Methods("DELETE")
}

func (h *Handler) handleCreateExperienceTech(w http.ResponseWriter, r *http.Request) {
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

	payload := types.ExperienceTechPayload{
		ExperienceID: experienceID,
		TechID:       techID,
	}

	if err := utils.Validate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid payload %v", errors))
		return
	}

	err = h.experienceTechStore.CreateExperienceTech(types.ExperienceTech{
		ExperienceID: payload.ExperienceID,
		TechID:       payload.TechID,
	})
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, ExperienceTechResponse{
		Message: "Technology successfully linked to experience",
	})
}

func (h *Handler) handleBatchCreateExperienceTech(w http.ResponseWriter, r *http.Request) {
	expId, err := utils.GetRequestId(r)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	var payload types.BatchExperienceTechPayload

	if err := utils.ParseJSON(r, &payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	if err := utils.Validate.Struct(payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	if err := h.experienceTechStore.CreateExperienceTechBatch(expId, payload.TechIDs); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, ExperienceTechResponse{
		Message: "Technologies successfully linked to experience",
	})
}

func (h *Handler) handleDeleteExperienceTech(w http.ResponseWriter, r *http.Request) {
	id, err := utils.GetRequestId(r)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	err = h.experienceTechStore.DeleteExperienceTech(id)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, ExperienceTechResponse{
		Message: "Experience tech successfully deleted",
	})
}
