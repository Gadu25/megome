package handler

import (
	"megome/internal/domain/apilog"
	"megome/internal/domain/personalaccesstoken"
	"megome/internal/domain/user"
	"megome/internal/middleware"
	"megome/internal/pkg/httputil"
	"net/http"

	"github.com/gorilla/mux"
)

type DashboardHandler struct {
	userStore        *user.Repository
	patStore         *personalaccesstoken.Repository
	apiUsageLogStore *apilog.Repository
}

type DasboardData struct {
	APIUsage apilog.UserAPIUsageStats `json:"apiUsage"`
	PATCount int                      `json:"patCount"`
}

type DashboardResponse struct {
	Message string       `json:"message"`
	Data    DasboardData `json:"data"`
}

func NewDashboardHandler(userStore *user.Repository, patStore *personalaccesstoken.Repository, apiUsageLogStore *apilog.Repository) *DashboardHandler {
	return &DashboardHandler{userStore: userStore, patStore: patStore, apiUsageLogStore: apiUsageLogStore}
}

func (h *DashboardHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/dashboard/overview", middleware.WithJWTAuth(h.handleViewDasboardOverview, h.userStore))
}

func (h *DashboardHandler) handleViewDasboardOverview(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r.Context())
	stats, err := h.apiUsageLogStore.GetUserUsageStats(userID)

	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	count, err := h.patStore.GetTokenCountByUserID(userID)

	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	dashData := DasboardData{
		APIUsage: stats,
		PATCount: count,
	}

	resp := DashboardResponse{
		Message: "Dashboard overview data",
		Data:    dashData,
	}

	httputil.WriteJSON(w, http.StatusOK, resp)
}
