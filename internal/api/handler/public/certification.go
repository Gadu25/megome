package public

import (
	"megome/internal/domain/apilog"
	"megome/internal/domain/certification"
	"megome/internal/domain/personalaccesstoken"
	"megome/internal/middleware"
	"megome/internal/pkg/httputil"
	"net/http"

	"github.com/gorilla/mux"
)

type CertificationHandler struct {
	certificationStore *certification.Repository
	patStore           *personalaccesstoken.Repository
	apiLogStore        *apilog.Repository
}

type CertificationPublicResponse struct {
	Message      string                       `json:"message"`
	Certificates []certification.Certification `json:"certificates"`
}

type SingleCertPublicResponse struct {
	Message     string                     `json:"message"`
	Certificate certification.Certification `json:"certificate"`
}

func NewCertificationHandler(certificationStore *certification.Repository, patStore *personalaccesstoken.Repository, apiLogStore *apilog.Repository) *CertificationHandler {
	return &CertificationHandler{
		certificationStore: certificationStore,
		patStore:           patStore,
		apiLogStore:        apiLogStore,
	}
}

func (h *CertificationHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/certification",
		middleware.WithPATAuth(
			middleware.WithAPILogging(h.handleGetPublicCertification, h.apiLogStore),
			h.patStore,
		),
	).Methods("GET")
	router.HandleFunc("/certification/{id}",
		middleware.WithPATAuth(
			middleware.WithAPILogging(h.handleGetPublicCertificationById, h.apiLogStore),
			h.patStore,
		),
	).Methods("GET")
}

func (h *CertificationHandler) handleGetPublicCertification(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetPATUserIDFromContext(r.Context())

	certificates, err := h.certificationStore.GetCertifications(userID)

	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	resp := CertificationPublicResponse{
		Message:      "certificates successfully fetched",
		Certificates: certificates,
	}

	httputil.WriteJSON(w, http.StatusOK, resp)
}

func (h *CertificationHandler) handleGetPublicCertificationById(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.GetRequestId(r)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	cert, err := h.certificationStore.GetCertificationById(id)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	resp := SingleCertPublicResponse{
		Message:     "certification fetched successfully",
		Certificate: cert,
	}
	httputil.WriteJSON(w, http.StatusOK, resp)
}
