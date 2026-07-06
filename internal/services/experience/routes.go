package experience

import (
	"bytes"
	"fmt"
	"io"
	"megome/internal/services/auth"
	"megome/internal/services/storage"
	"megome/internal/services/types"
	"megome/internal/services/utils"
	"net/http"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
)

type Handler struct {
	experienceStore types.ExperienceStore
	userStore       types.UserStore
	r2Client   *storage.R2Client
}

type ExperienceResponse struct {
	Message     string             `json:"message"`
	Experiences []types.Experience `json:"experiences"`
}

type SingleExpResponse struct {
	Message    string           `json:"message"`
	Experience types.Experience `json:"experience"`
}

func NewHandler(experienceStore types.ExperienceStore, userStore types.UserStore, r2Client *storage.R2Client) *Handler {
	return &Handler{experienceStore: experienceStore, userStore: userStore, r2Client: r2Client}
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
		Message:     "Experience fetched successfully",
		Experiences: experiences,
	}
	utils.WriteJSON(w, http.StatusOK, resp)
}

func (h *Handler) uploadLogo(r *http.Request) (*string, error) {
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

	sniffLen := 512
	if len(data) < sniffLen {
		sniffLen = len(data)
	}

	fileType := http.DetectContentType(data[:sniffLen])
	if fileType != "image/jpeg" && fileType != "image/png" && fileType != "image/webp" {
		return nil, fmt.Errorf("invalid file type")
	}

	key, err := storage.GenerateKey(
		fmt.Sprintf("experience/%d/logo", auth.GetUserIDFromContext(r.Context())),
		utils.GenerateUUID(),
		fileType,
	)
	if err != nil {
		return nil, err
	}

	err = h.r2Client.UploadFromReader(
		r.Context(),
		key,
		bytes.NewReader(data),
		int64(len(data)),
		handler.Header.Get("Content-Type"),
	)
	if err != nil {
		return nil, err
	}

	return &key, nil
}

func (h *Handler) handleCreateExperience(w http.ResponseWriter, r *http.Request) {
	isPresent, err := strconv.ParseBool(r.FormValue("isPresent"))
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("Error handling payload: %w", err))
	}

	payload := types.ExperiencePayload{
		Title: r.FormValue("title"),
		Company: r.FormValue("company"),
		StartDate: r.FormValue("startDate"),
		EndDate: utils.PointerFromString(r.FormValue("endDate")),
		IsPresent: isPresent,
		Description: r.FormValue("description"),
	}

	if err := utils.Validate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid payload %v", errors))
		return
	}

	userID := auth.GetUserIDFromContext(r.Context())

	logoKey, err := h.uploadLogo(r)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	exp, err := h.experienceStore.CreateExperience(types.Experience{
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
	isPresent, err := strconv.ParseBool(r.FormValue("isPresent"))
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("Error handling payload: %w", err))
	}

	payload := types.ExperiencePayload{
		Title:       r.FormValue("title"),
		Company:     r.FormValue("company"),
		StartDate:   r.FormValue("startDate"),
		EndDate:     utils.PointerFromString(r.FormValue("endDate")),
		IsPresent:   isPresent,
		Description: r.FormValue("description"),
	}

	if err := utils.Validate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid payload %v", errors))
		return
	}

	id, err := utils.GetRequestId(r)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	existing, err := h.experienceStore.GetExperienceById(id)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	logoKey, err := h.uploadLogo(r)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	if logoKey != nil && existing.Logo != nil {
		_ = h.r2Client.DeleteObject(r.Context(), *existing.Logo)
	}

	exp, err := h.experienceStore.UpdateExperience(id, types.Experience{
		Title:       payload.Title,
		Company:     payload.Company,
		Logo:        logoKey,
		StartDate:   payload.StartDate,
		EndDate:     payload.EndDate,
		IsPresent:   payload.IsPresent,
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
