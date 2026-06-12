package dashboard

import (
	"megome/internal/services/auth"
	"megome/internal/services/types"
	"megome/internal/services/utils"
	"net/http"

	"github.com/gorilla/mux"
)

type Handler struct {
	userStore        types.UserStore
	patStore         types.PersonalAccessTokenStore
	apiUsageLogStore types.APIUsageLogStore
}

type DasboardData struct {
	APIUsage types.UserAPIUsageStats `json:"apiUsage"`
	PATCount int                     `json:"patCount"`
}

type Response struct {
	Message string       `json:"message"`
	Data    DasboardData `json:"data"`
}

func NewHandler(userStore types.UserStore, patStore types.PersonalAccessTokenStore, apiUsageLogStore types.APIUsageLogStore) *Handler {
	return &Handler{userStore: userStore, patStore: patStore, apiUsageLogStore: apiUsageLogStore}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/dashboard/overview", auth.WithJWTAuth(h.handleViewDasboardOverview, h.userStore))
}

func (h *Handler) handleViewDasboardOverview(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetPATTokenIDFromContext(r.Context())
	stats, err := h.apiUsageLogStore.GetUserUsageStats(userID)

	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	count, err := h.patStore.GetTokenCountByUserID(userID)

	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	dashData := DasboardData{
		APIUsage: stats,
		PATCount: count,
	}

	resp := Response{
		Message: "Dashboard overview data",
		Data:    dashData,
	}

	utils.WriteJSON(w, http.StatusOK, resp)
}
