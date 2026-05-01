package experience

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
	experienceStore types.ExperienceStore
	userStore       types.UserStore
}

type ExperienceResponse struct {
	Message    string             `json:"message"`
	Experience []types.Experience `json:"experience"`
}

type SingleExpResponse struct {
	Message    string           `json:"message"`
	Experience types.Experience `json:"experience"`
}

func NewHandler(experienceStore types.ExperienceStore, userStore types.UserStore) *Handler {
	return &Handler{experienceStore: experienceStore, userStore: userStore}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/experience", auth.WithJWTAuth(h.handleViewExperiences, h.userStore)).Methods("GET")
	router.HandleFunc("/experience", auth.WithJWTAuth(h.handleCreateExperience, h.userStore)).Methods("POST")
	router.HandleFunc("/experience/{id}", auth.WithJWTAuth(h.handleEditExperience, h.userStore)).Methods("PUT")
	router.HandleFunc("/experience/{id}", auth.WithJWTAuth(h.handleDeleteExperience, h.userStore)).Methods("DELETE")
}

func (h *Handler) handleViewExperiences(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserIDFromContext(r.Context())
	experiences, err := h.experienceStore.GetExperiences(userID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}
	resp := ExperienceResponse{
		Message:    "Experience fetched successfully",
		Experience: experiences,
	}
	utils.WriteJSON(w, http.StatusOK, resp)
}

func (h *Handler) handleCreateExperience(w http.ResponseWriter, r *http.Request) {
	// get JSON payload
	endDateStr := r.FormValue("endDate")

	var endDate *string
	if endDateStr != "" {
		endDate = &endDateStr
	}

	payload := types.ExperiencePayload{
		Title:       r.FormValue("title"),
		Company:     r.FormValue("company"),
		StartDate:   r.FormValue("startDate"),
		EndDate:     endDate,
		Description: r.FormValue("description"),
	}

	// validate the payload
	if err := utils.Validate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid payload %v", errors))
		return
	}

	// create experience
	userID := auth.GetUserIDFromContext(r.Context())
	exp, err := h.experienceStore.CreateExperience(types.Experience{
		UserID:      userID,
		Title:       payload.Title,
		Company:     payload.Company,
		StartDate:   payload.StartDate,
		EndDate:     payload.EndDate,
		Description: payload.Description,
	})
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	resp := SingleExpResponse{
		Message:    "Experience created successfully",
		Experience: exp,
	}
	utils.WriteJSON(w, http.StatusOK, resp)
}

func (h *Handler) handleEditExperience(w http.ResponseWriter, r *http.Request) {
	endDateStr := r.FormValue("endDate")

	var endDate *string
	if endDateStr != "" {
		endDate = &endDateStr
	}

	payload := types.ExperiencePayload{
		Title:       r.FormValue("title"),
		Company:     r.FormValue("company"),
		StartDate:   r.FormValue("startDate"),
		EndDate:     endDate,
		Description: r.FormValue("description"),
	}

	// validate the payload
	if err := utils.Validate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid payload %v", errors))
		return
	}

	// update experience
	id, err := utils.GetRequestId(r)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}
	exp, err := h.experienceStore.UpdateExperience(id, types.Experience{
		Title:       payload.Title,
		Company:     payload.Company,
		StartDate:   payload.StartDate,
		EndDate:     payload.EndDate,
		Description: payload.Description,
	})
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	resp := SingleExpResponse{
		Message:    "Experience updated successfully",
		Experience: exp,
	}
	utils.WriteJSON(w, http.StatusOK, resp)
}

func (h *Handler) handleDeleteExperience(w http.ResponseWriter, r *http.Request) {
	id, err := utils.GetRequestId(r)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}
	exp, err := h.experienceStore.DeleteExperience(id)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	resp := SingleExpResponse{
		Message:    "Experience deleted successfully",
		Experience: exp,
	}
	utils.WriteJSON(w, http.StatusOK, resp)
}
