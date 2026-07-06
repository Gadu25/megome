package certification

import (
	"bytes"
	"fmt"
	"io"
	"megome/internal/services/auth"
	"megome/internal/services/storage"
	"megome/internal/services/types"
	"megome/internal/services/utils"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
)

type Handler struct {
	certificationStore types.CertificationStore
	userStore          types.UserStore
	r2Client           *storage.R2Client
}

type CertificationResponse struct {
	Message      string                `json:"message"`
	Certificates []types.Certification `json:"certificates"`
}

type SingleCertResponse struct {
	Message     string              `json:"message"`
	Certificate types.Certification `json:"certificate"`
}

func NewHandler(certificationStore types.CertificationStore, userStore types.UserStore, r2Client *storage.R2Client) *Handler {
	return &Handler{certificationStore: certificationStore, userStore: userStore, r2Client: r2Client}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/certification", auth.WithJWTAuth(h.handleViewCertification, h.userStore)).Methods("GET")
	router.HandleFunc("/certification", auth.WithJWTAuth(h.handleCreateCertification, h.userStore)).Methods("POST")
	router.HandleFunc("/certification/{id}", auth.WithJWTAuth(h.handleEditCertification, h.userStore)).Methods("PUT")
	router.HandleFunc("/certification/{id}", auth.WithJWTAuth(h.handleDeleteCertification, h.userStore)).Methods("DELETE")
}

func (h *Handler) uploadCertificateImage(r *http.Request) (*string, error) {
	file, handler, err := r.FormFile("certificateImage")
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
		fmt.Sprintf("certification/%d/certificateImage", auth.GetUserIDFromContext(r.Context())),
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

func (h *Handler) handleViewCertification(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserIDFromContext(r.Context())
	certifications, err := h.certificationStore.GetCertifications(userID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}
	resp := CertificationResponse{
		Message:      "Certification fetched successfully",
		Certificates: certifications,
	}
	utils.WriteJSON(w, http.StatusOK, resp)
}

func (h *Handler) handleCreateCertification(w http.ResponseWriter, r *http.Request) {
	payload := types.CertificationPayload{
		Title:          r.FormValue("title"),
		Issuer:         r.FormValue("issuer"),
		IssueDate:      r.FormValue("issueDate"),
		ExpirationDate: utils.PointerFromString(r.FormValue("expirationDate")),
		CredentialId:   utils.PointerFromString(r.FormValue("credentialId")),
		CredentialUrl:  utils.PointerFromString(r.FormValue("credentialUrl")),
	}

	if err := utils.Validate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid payload %v", errors))
		return
	}

	userID := auth.GetUserIDFromContext(r.Context())

	certificateImageKey, err := h.uploadCertificateImage(r)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	cert, err := h.certificationStore.CreateCertification(types.Certification{
		UserID:           userID,
		Title:            payload.Title,
		Issuer:           payload.Issuer,
		IssueDate:        payload.IssueDate,
		CertificateImage: certificateImageKey,
		ExpirationDate:   payload.ExpirationDate,
		CredentialId:     payload.CredentialId,
		CredentialUrl:    payload.CredentialUrl,
	})

	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	resp := SingleCertResponse{
		Message:     "Certification created successfully",
		Certificate: cert,
	}
	utils.WriteJSON(w, http.StatusOK, resp)
}
func (h *Handler) handleEditCertification(w http.ResponseWriter, r *http.Request) {
	payload := types.CertificationPayload{
		Title:          r.FormValue("title"),
		Issuer:         r.FormValue("issuer"),
		IssueDate:      r.FormValue("issueDate"),
		ExpirationDate: utils.PointerFromString(r.FormValue("expirationDate")),
		CredentialId:   utils.PointerFromString(r.FormValue("credentialId")),
		CredentialUrl:  utils.PointerFromString(r.FormValue("credentialUrl")),
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

	existing, err := h.certificationStore.GetCertificationById(id)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	certificateImageKey, err := h.uploadCertificateImage(r)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	if certificateImageKey != nil && existing.CertificateImage != nil {
		_ = h.r2Client.DeleteObject(r.Context(), *existing.CertificateImage)
	}

	cert, err := h.certificationStore.UpdateCertification(id, types.Certification{
		Title:            payload.Title,
		Issuer:           payload.Issuer,
		IssueDate:        payload.IssueDate,
		CertificateImage: certificateImageKey,
		ExpirationDate:   payload.ExpirationDate,
		CredentialId:     payload.CredentialId,
		CredentialUrl:    payload.CredentialUrl,
	})
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	resp := SingleCertResponse{
		Message:     "Certification edited successfully",
		Certificate: cert,
	}
	utils.WriteJSON(w, http.StatusOK, resp)
}
func (h *Handler) handleDeleteCertification(w http.ResponseWriter, r *http.Request) {
	id, err := utils.GetRequestId(r)
	fmt.Println("[DEBUG]", id)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	cert, err := h.certificationStore.DeleteCertification(id)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	if cert.CertificateImage != nil && *cert.CertificateImage != "" {
		_ = h.r2Client.DeleteObject(r.Context(), *cert.CertificateImage)
	}

	resp := SingleCertResponse{
		Message:     "Certification deleted successfully",
		Certificate: cert,
	}
	utils.WriteJSON(w, http.StatusOK, resp)
}
