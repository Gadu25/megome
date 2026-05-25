package publiccertification

import (
	"megome/internal/platform/http/middleware"
	"megome/internal/services/auth"
	"megome/internal/services/types"
	"megome/internal/services/utils"
	"net/http"

	"github.com/gorilla/mux"
)

type Handler struct {
	certificationStore types.CertificationStore
	patStore           types.PersonalAccessTokenStore
	apiLogStore        types.APIUsageLogStore
}

type PublicResponse struct {
	Message      string                `json:"message"`
	Certificates []types.Certification `json:"certificates"`
}

func NewHandler(certificationStore types.CertificationStore, patStore types.PersonalAccessTokenStore, apiLogStore types.APIUsageLogStore) *Handler {
	return &Handler{
		certificationStore: certificationStore,
		patStore:           patStore,
		apiLogStore:        apiLogStore,
	}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/certificate",
		auth.WithPATAuth(
			middleware.WithAPILogging(h.handleGetPublicSkill, h.apiLogStore),
			h.patStore,
		),
	).Methods("GET")
}

func (h *Handler) handleGetPublicSkill(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetPATUserIDFromContext(r.Context())

	certificates, err := h.certificationStore.GetPublicCertifications(userID)

	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	resp := PublicResponse{
		Message:      "certificates successfully fetched",
		Certificates: certificates,
	}

	utils.WriteJSON(w, http.StatusOK, resp)
}
