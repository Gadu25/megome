package certification

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
	certificationStore types.CertificationStore
	userStore          types.UserStore
}

type CertificationResponse struct {
	Message string                `json:"message"`
	Data    []types.Certification `json:"data"`
}

func NewHandler(certificationStore types.CertificationStore, userStore types.UserStore) *Handler {
	return &Handler{certificationStore: certificationStore, userStore: userStore}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/certification", auth.WithJWTAuth(h.handleViewCertification, h.userStore)).Methods("GET")
	router.HandleFunc("/certification", auth.WithJWTAuth(h.handleCreateCertification, h.userStore)).Methods("POST")
	router.HandleFunc("/certification/{id}", auth.WithJWTAuth(h.handleEditCertification, h.userStore)).Methods("PUT")
	router.HandleFunc("/certification/{id}", auth.WithJWTAuth(h.handleDeleteCertification, h.userStore)).Methods("DELETE")
}

func (h *Handler) handleViewCertification(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserIDFromContext(r.Context())
	certifications, err := h.certificationStore.GetCertifications(userID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}
	resp := CertificationResponse{
		Message: "Certification fetched successfully",
		Data:    certifications,
	}
	utils.WriteJSON(w, http.StatusOK, resp)
}

func (h *Handler) handleCreateCertification(w http.ResponseWriter, r *http.Request) {
	// get JSON payload
	var payload types.CertificationPayload
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
	userID := auth.GetUserIDFromContext(r.Context())
	err := h.certificationStore.CreateCertification(types.Certification{
		UserID:         userID,
		Title:          payload.Title,
		Issuer:         payload.Issuer,
		IssueDate:      payload.IssueDate,
		ExpirationDate: payload.ExpirationDate,
		CredentialId:   payload.CredentialId,
		CredentialUrl:  payload.CredentialUrl,
	})
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}
	utils.WriteJSON(w, http.StatusOK, map[string]string{
		"message": "Certification is successfully created",
	})
}
func (h *Handler) handleEditCertification(w http.ResponseWriter, r *http.Request) {
	// get JSON payload
	var payload types.CertificationPayload
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

	// update certification
	id, err := utils.GetRequestId(r)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}
	err = h.certificationStore.UpdateCertification(id, types.Certification{
		Title:          payload.Title,
		Issuer:         payload.Issuer,
		IssueDate:      payload.IssueDate,
		ExpirationDate: payload.ExpirationDate,
		CredentialId:   payload.CredentialId,
		CredentialUrl:  payload.CredentialUrl,
	})
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}
	utils.WriteJSON(w, http.StatusOK, map[string]string{
		"message": "Certification is successfully updated",
	})
}
func (h *Handler) handleDeleteCertification(w http.ResponseWriter, r *http.Request) {
	id, err := utils.GetRequestId(r)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}
	err = h.certificationStore.DeleteCertification(id)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}
	utils.WriteJSON(w, http.StatusOK, map[string]string{
		"message": "Certification is successfully deleted",
	})
}
