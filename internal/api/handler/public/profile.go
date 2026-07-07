package public

import (
	"megome/internal/domain/apilog"
	"megome/internal/domain/personalaccesstoken"
	"megome/internal/domain/profile"
	"megome/internal/middleware"
	"megome/internal/pkg/httputil"
	"net/http"

	"github.com/gorilla/mux"
)

type ProfileHandler struct {
	profileStore *profile.Repository
	patStore     *personalaccesstoken.Repository
	apiLogStore  *apilog.Repository
}

type PublicResponse struct {
	Message string          `json:"message"`
	Data    *profile.Profile `json:"data"`
}

func NewProfileHandler(profileStore *profile.Repository, patStore *personalaccesstoken.Repository, apiLogStore *apilog.Repository) *ProfileHandler {
	return &ProfileHandler{
		profileStore: profileStore,
		patStore:     patStore,
		apiLogStore:  apiLogStore,
	}
}

func (h *ProfileHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/profile",
		middleware.WithPATAuth(
			middleware.WithAPILogging(h.handleGetPublicProfile, h.apiLogStore),
			h.patStore,
		),
	).Methods("GET")
}

func (h *ProfileHandler) handleGetPublicProfile(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetPATUserIDFromContext(r.Context())

	p, err := h.profileStore.GetPublicProfile(userID)

	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	resp := PublicResponse{
		Message: "profile successfully fetched",
		Data:    p,
	}

	httputil.WriteJSON(w, http.StatusOK, resp)
}
