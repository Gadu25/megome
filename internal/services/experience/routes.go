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
	Message string             `json:"message"`
	Data    []types.Experience `json:"data"`
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
		Message: "Experience fetched successfully",
		Data:    experiences,
	}
	utils.WriteJSON(w, http.StatusOK, resp)
}

func (h *Handler) handleCreateExperience(w http.ResponseWriter, r *http.Request) {
	// get JSON payload
	var payload types.ExperiencePayload
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

	// create experience
	userID := auth.GetUserIDFromContext(r.Context())
	err := h.experienceStore.CreateExperience(types.Experience{
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
	utils.WriteJSON(w, http.StatusOK, map[string]string{"message": "Experience is successfully created"})
}

func (h *Handler) handleEditExperience(w http.ResponseWriter, r *http.Request) {
	// body, _ := io.ReadAll(r.Body)
	// fmt.Println(string(body)) // check exactly what the server received
	// get JSON payload
	var payload types.ExperiencePayload
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

	// update experience
	id, err := utils.GetRequestId(r)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}
	err = h.experienceStore.UpdateExperience(id, types.Experience{
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
	utils.WriteJSON(w, http.StatusOK, map[string]string{"message": "Experience is successfully updated"})
}

func (h *Handler) handleDeleteExperience(w http.ResponseWriter, r *http.Request) {
	id, err := utils.GetRequestId(r)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}
	err = h.experienceStore.DeleteExperience(id)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}
	utils.WriteJSON(w, http.StatusOK, map[string]string{"message": "Experience is successfully deleted"})
}
