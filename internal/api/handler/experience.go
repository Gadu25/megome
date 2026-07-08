package handler

import (
	"fmt"
	"io"
	"megome/internal/domain/experience"
	"megome/internal/domain/user"
	"megome/internal/middleware"
	"megome/internal/pkg/httputil"
	"megome/internal/pkg/storage"
	"megome/internal/pkg/validator"
	"net/http"
	"strconv"

	playvalidator "github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
)

type ExperienceHandler struct {
	experienceStore *experience.Repository
	userStore       *user.Repository
	r2Client        *storage.R2Client
}

type ExperienceResponse struct {
	Message     string                `json:"message"`
	Experiences []experience.Experience `json:"experiences"`
}

type SingleExpResponse struct {
	Message    string              `json:"message"`
	Experience experience.Experience `json:"experience"`
}

func NewExperienceHandler(experienceStore *experience.Repository, userStore *user.Repository, r2Client *storage.R2Client) *ExperienceHandler {
	return &ExperienceHandler{experienceStore: experienceStore, userStore: userStore, r2Client: r2Client}
}

func (h *ExperienceHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/experience", middleware.WithJWTAuth(h.handleViewExperiences, h.userStore)).Methods("GET")
	router.HandleFunc("/experience", middleware.WithJWTAuth(h.handleCreateExperience, h.userStore)).Methods("POST")
	router.HandleFunc("/experience/{id}", middleware.WithJWTAuth(h.handleEditExperience, h.userStore)).Methods("PUT")
	router.HandleFunc("/experience/{id}", middleware.WithJWTAuth(h.handleDeleteExperience, h.userStore)).Methods("DELETE")
}

func (h *ExperienceHandler) handleViewExperiences(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r.Context())
	experiences, err := h.experienceStore.GetExperiences(userID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err)
		return
	}
	resp := ExperienceResponse{
		Message:     "Experience fetched successfully",
		Experiences: experiences,
	}
	httputil.WriteJSON(w, http.StatusOK, resp)
}

func (h *ExperienceHandler) uploadLogo(r *http.Request) (*string, error) {
	file, handler, err := r.FormFile("logo")
	if err != nil {
		return nil, nil
	}
	defer file.Close()

	if handler.Size > 1<<20 {
		return nil, fmt.Errorf("file too large (max 1MB)")
	}

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	key, err := h.r2Client.UploadImage(r.Context(), data,
		fmt.Sprintf("experience/%d/logo", middleware.GetUserIDFromContext(r.Context())),
		httputil.GenerateUUID(),
	)
	if err != nil {
		return nil, err
	}

	return &key, nil
}

func (h *ExperienceHandler) handleCreateExperience(w http.ResponseWriter, r *http.Request) {
	isPresent, err := strconv.ParseBool(r.FormValue("isPresent"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, fmt.Errorf("Error handling payload: %w", err))
	}

	payload := experience.ExperiencePayload{
		Title:       r.FormValue("title"),
		Company:     r.FormValue("company"),
		StartDate:   r.FormValue("startDate"),
		EndDate:     httputil.PointerFromString(r.FormValue("endDate")),
		IsPresent:   isPresent,
		Description: r.FormValue("description"),
	}

	if err := validator.Validate.Struct(payload); err != nil {
		errors := err.(playvalidator.ValidationErrors)
		httputil.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid payload %v", errors))
		return
	}

	userID := middleware.GetUserIDFromContext(r.Context())

	logoKey, err := h.uploadLogo(r)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	exp, err := h.experienceStore.CreateExperience(experience.Experience{
		UserID:      userID,
		Title:       payload.Title,
		Company:     payload.Company,
		Logo:        logoKey,
		StartDate:   payload.StartDate,
		EndDate:     payload.EndDate,
		IsPresent:   payload.IsPresent,
		Description: payload.Description,
	})

	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	resp := SingleExpResponse{
		Message:    "Experience created successfully",
		Experience: exp,
	}
	httputil.WriteJSON(w, http.StatusOK, resp)
}

func (h *ExperienceHandler) handleEditExperience(w http.ResponseWriter, r *http.Request) {
	isPresent, err := strconv.ParseBool(r.FormValue("isPresent"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, fmt.Errorf("Error handling payload: %w", err))
	}

	payload := experience.ExperiencePayload{
		Title:       r.FormValue("title"),
		Company:     r.FormValue("company"),
		StartDate:   r.FormValue("startDate"),
		EndDate:     httputil.PointerFromString(r.FormValue("endDate")),
		IsPresent:   isPresent,
		Description: r.FormValue("description"),
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

	existing, err := h.experienceStore.GetExperienceById(id)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	logoKey, err := h.uploadLogo(r)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	if logoKey != nil && existing.Logo != nil {
		_ = h.r2Client.DeleteObject(r.Context(), *existing.Logo)
	}

	exp, err := h.experienceStore.UpdateExperience(id, experience.Experience{
		Title:       payload.Title,
		Company:     payload.Company,
		Logo:        logoKey,
		StartDate:   payload.StartDate,
		EndDate:     payload.EndDate,
		IsPresent:   payload.IsPresent,
		Description: payload.Description,
	})
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	resp := SingleExpResponse{
		Message:    "Experience updated successfully",
		Experience: exp,
	}
	httputil.WriteJSON(w, http.StatusOK, resp)
}

func (h *ExperienceHandler) handleDeleteExperience(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.GetRequestId(r)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err)
		return
	}
	exp, err := h.experienceStore.DeleteExperience(id)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	resp := SingleExpResponse{
		Message:    "Experience deleted successfully",
		Experience: exp,
	}
	httputil.WriteJSON(w, http.StatusOK, resp)
}
