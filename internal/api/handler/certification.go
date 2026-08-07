package handler

import (
	"fmt"
	"io"
	"megome/internal/domain/certification"
	"megome/internal/domain/user"
	"megome/internal/middleware"
	"megome/internal/pkg/httputil"
	"megome/internal/pkg/storage"
	"megome/internal/pkg/validator"
	"net/http"

	playvalidator "github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
)

type CertificationHandler struct {
	certificationStore *certification.Repository
	userStore          *user.Repository
	r2Client           *storage.R2Client
}

type CertificationResponse struct {
	Message      string                  `json:"message"`
	Certificates []certification.Certification `json:"certificates"`
}

type SingleCertResponse struct {
	Message     string                    `json:"message"`
	Certificate certification.Certification `json:"certificate"`
}

func NewCertificationHandler(certificationStore *certification.Repository, userStore *user.Repository, r2Client *storage.R2Client) *CertificationHandler {
	return &CertificationHandler{certificationStore: certificationStore, userStore: userStore, r2Client: r2Client}
}

func (h *CertificationHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/certification", middleware.WithJWTAuth(h.handleViewCertification, h.userStore)).Methods("GET")
	router.HandleFunc("/certification/reorder", middleware.WithJWTAuth(h.handleReorderCertifications, h.userStore)).Methods("POST")
	router.HandleFunc("/certification/{id}", middleware.WithJWTAuth(h.handleViewCertificationById, h.userStore)).Methods("GET")
	router.HandleFunc("/certification", middleware.WithJWTAuth(h.handleCreateCertification, h.userStore)).Methods("POST")
	router.HandleFunc("/certification/{id}", middleware.WithJWTAuth(h.handleEditCertification, h.userStore)).Methods("PUT")
	router.HandleFunc("/certification/{id}", middleware.WithJWTAuth(h.handleDeleteCertification, h.userStore)).Methods("DELETE")
}

func (h *CertificationHandler) uploadCertificateImage(r *http.Request) (*string, error) {
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

	key, err := h.r2Client.UploadImage(r.Context(), data,
		fmt.Sprintf("certification/%d/certificateImage", middleware.GetUserIDFromContext(r.Context())),
		httputil.GenerateUUID(),
	)
	if err != nil {
		return nil, err
	}

	return &key, nil
}

func (h *CertificationHandler) handleViewCertification(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r.Context())

	limit := 20
	offset := 0
	query := r.URL.Query()
	if l := query.Get("limit"); l != "" {
		limit = httputil.ParseIntOrDefault(l, 20)
	}
	if o := query.Get("offset"); o != "" {
		offset = httputil.ParseIntOrDefault(o, 0)
	}
	if limit > 100 {
		limit = 100
	}

	certifications, err := h.certificationStore.GetCertifications(userID, limit, offset)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	total, err := h.certificationStore.CountByUserID(userID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	resp := certification.PaginatedCertificationsResponse{
		Data: certifications,
	}
	resp.Pagination.Limit = limit
	resp.Pagination.Offset = offset
	resp.Pagination.Total = total
	httputil.WriteJSON(w, http.StatusOK, resp)
}

func (h *CertificationHandler) handleViewCertificationById(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.GetRequestId(r)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err)
		return
	}

	cert, err := h.certificationStore.GetCertificationById(id)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	resp := SingleCertResponse{
		Message:     "Certification fetched successfully",
		Certificate: cert,
	}
	httputil.WriteJSON(w, http.StatusOK, resp)
}

func (h *CertificationHandler) handleCreateCertification(w http.ResponseWriter, r *http.Request) {
	payload := certification.CertificationPayload{
		Title:          r.FormValue("title"),
		Issuer:         r.FormValue("issuer"),
		IssueDate:      r.FormValue("issueDate"),
		ExpirationDate: httputil.PointerFromString(r.FormValue("expirationDate")),
		CredentialId:   httputil.PointerFromString(r.FormValue("credentialId")),
		CredentialUrl:  httputil.PointerFromString(r.FormValue("credentialUrl")),
	}

	if err := validator.Validate.Struct(payload); err != nil {
		errors := err.(playvalidator.ValidationErrors)
		httputil.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid payload %v", errors))
		return
	}

	userID := middleware.GetUserIDFromContext(r.Context())

	certificateImageKey, err := h.uploadCertificateImage(r)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	cert, err := h.certificationStore.CreateCertification(certification.Certification{
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
		httputil.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	resp := SingleCertResponse{
		Message:     "Certification created successfully",
		Certificate: cert,
	}
	httputil.WriteJSON(w, http.StatusOK, resp)
}

func (h *CertificationHandler) handleEditCertification(w http.ResponseWriter, r *http.Request) {
	payload := certification.CertificationPayload{
		Title:          r.FormValue("title"),
		Issuer:         r.FormValue("issuer"),
		IssueDate:      r.FormValue("issueDate"),
		ExpirationDate: httputil.PointerFromString(r.FormValue("expirationDate")),
		CredentialId:   httputil.PointerFromString(r.FormValue("credentialId")),
		CredentialUrl:  httputil.PointerFromString(r.FormValue("credentialUrl")),
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

	existing, err := h.certificationStore.GetCertificationById(id)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	certificateImageKey, err := h.uploadCertificateImage(r)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	if certificateImageKey != nil && existing.CertificateImage != nil {
		_ = h.r2Client.DeleteObject(r.Context(), *existing.CertificateImage)
	}

	finalImage := certificateImageKey
	if finalImage == nil && existing.CertificateImage != nil {
		key := httputil.ExtractR2Key(*existing.CertificateImage)
		if key != "" {
			finalImage = &key
		}
	}

	cert, err := h.certificationStore.UpdateCertification(id, certification.Certification{
		Title:            payload.Title,
		Issuer:           payload.Issuer,
		IssueDate:        payload.IssueDate,
		CertificateImage: finalImage,
		ExpirationDate:   payload.ExpirationDate,
		CredentialId:     payload.CredentialId,
		CredentialUrl:    payload.CredentialUrl,
	})
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	resp := SingleCertResponse{
		Message:     "Certification edited successfully",
		Certificate: cert,
	}
	httputil.WriteJSON(w, http.StatusOK, resp)
}

func (h *CertificationHandler) handleDeleteCertification(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.GetRequestId(r)
	fmt.Println("[DEBUG]", id)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err)
		return
	}

	userID := middleware.GetUserIDFromContext(r.Context())
	cert, err := h.certificationStore.DeleteCertification(id, userID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	if cert.CertificateImage != nil && *cert.CertificateImage != "" {
		_ = h.r2Client.DeleteObject(r.Context(), *cert.CertificateImage)
	}

	resp := SingleCertResponse{
		Message:     "Certification deleted successfully",
		Certificate: cert,
	}
	httputil.WriteJSON(w, http.StatusOK, resp)
}

func (h *CertificationHandler) handleReorderCertifications(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Items []certification.ReorderItem `json:"items"`
	}
	if err := httputil.ParseJSON(r, &payload); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err)
		return
	}

	if len(payload.Items) == 0 {
		httputil.WriteError(w, http.StatusBadRequest, fmt.Errorf("items list cannot be empty"))
		return
	}

	seen := make(map[int]bool, len(payload.Items))
	for _, item := range payload.Items {
		if seen[item.ID] {
			httputil.WriteError(w, http.StatusBadRequest, fmt.Errorf("duplicate item id: %d", item.ID))
			return
		}
		seen[item.ID] = true
	}

	userID := middleware.GetUserIDFromContext(r.Context())

	if err := h.certificationStore.ReorderCertifications(userID, payload.Items); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	httputil.WriteJSON(w, http.StatusOK, map[string]string{"message": "Certifications reordered successfully"})
}
