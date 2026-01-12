package education

import (
	"fmt"
	"megome/internal/services/auth"
	"megome/internal/services/types"
	"megome/internal/services/utils"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
)

type Handler struct {
	educationStore types.EducationStore
	userStore      types.UserStore
}

type EducationResponse struct {
	Message string            `json:"message"`
	Data    []types.Education `json:"data"`
}

func NewHandler(educationStore types.EducationStore, userStore types.UserStore) *Handler {
	return &Handler{educationStore: educationStore, userStore: userStore}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/education", auth.WithJWTAuth(h.handleViewEducation, h.userStore)).Methods("GET")
	router.HandleFunc("/education", auth.WithJWTAuth(h.handleCreateEducation, h.userStore)).Methods("POST")
	router.HandleFunc("/education/{id}", auth.WithJWTAuth(h.handleEditEducation, h.userStore)).Methods("PUT")
	router.HandleFunc("/education/{id}", auth.WithJWTAuth(h.handleDeleteEducation, h.userStore)).Methods("DELETE")
}

func (h *Handler) handleViewEducation(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserIDFromContext(r.Context())
	educations, err := h.educationStore.GetEducations(userID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}
	resp := EducationResponse{
		Message: "Education fetched successfully",
		Data:    educations,
	}
	utils.WriteJSON(w, http.StatusOK, resp)
}
func (h *Handler) handleCreateEducation(w http.ResponseWriter, r *http.Request) {
	// get JSON payload
	var payload types.EducationPayload
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

	// create education
	userID := auth.GetUserIDFromContext(r.Context())
	err := h.educationStore.CreateEducation(types.Education{
		UserID:       userID,
		School:       payload.School,
		FieldOfStudy: payload.FieldOfStudy,
		StartDate:    payload.StartDate,
		EndDate:      payload.EndDate,
	})
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}
	utils.WriteJSON(w, http.StatusOK, map[string]string{
		"message": "Education is successfully created",
	})
}
func (h *Handler) handleEditEducation(w http.ResponseWriter, r *http.Request) {
	// get JSON payload
	var payload types.EducationPayload
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

	// update education
	id, err := utils.GetRequestId(r)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}
	err = h.educationStore.UpdateEducation(id, types.Education{
		School:       payload.School,
		FieldOfStudy: payload.FieldOfStudy,
		StartDate:    payload.StartDate,
		EndDate:      payload.EndDate,
	})
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}
	utils.WriteJSON(w, http.StatusOK, map[string]string{"message": "Education is successfully updated"})
}
func (h *Handler) handleDeleteEducation(w http.ResponseWriter, r *http.Request) {
	id, err := utils.GetRequestId(r)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}
	err = h.educationStore.DeleteEducation(id)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}
	utils.WriteJSON(w, http.StatusOK, map[string]string{"message": "Education is successfully deleted"})
}
