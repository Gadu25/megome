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

type ActivityResponse struct {
	Message string                   `json:"message"`
	Data    []apilog.DashboardActivity `json:"data"`
}

type UsageStatsResponse struct {
	Message string              `json:"message"`
	Data    []apilog.DailyUsage `json:"data"`
}

func NewDashboardHandler(userStore *user.Repository, patStore *personalaccesstoken.Repository, apiUsageLogStore *apilog.Repository) *DashboardHandler {
	return &DashboardHandler{userStore: userStore, patStore: patStore, apiUsageLogStore: apiUsageLogStore}
}

func (h *DashboardHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/dashboard/overview", middleware.WithJWTAuth(h.handleViewDasboardOverview, h.userStore))
	router.HandleFunc("/dashboard/activity", middleware.WithJWTAuth(h.handleViewActivity, h.userStore)).Methods("GET")
	router.HandleFunc("/dashboard/usage-stats", middleware.WithJWTAuth(h.handleViewUsageStats, h.userStore)).Methods("GET")
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

func (h *DashboardHandler) handleViewActivity(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r.Context())

	limit := 20
	if l := r.URL.Query().Get("limit"); l != "" {
		limit = httputil.ParseIntOrDefault(l, 20)
	}

	activities, err := h.apiUsageLogStore.GetRecentActivity(userID, limit)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	resp := ActivityResponse{
		Message: "Recent activity fetched successfully",
		Data:    activities,
	}
	httputil.WriteJSON(w, http.StatusOK, resp)
}

func (h *DashboardHandler) handleViewUsageStats(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r.Context())

	days := 30
	if d := r.URL.Query().Get("days"); d != "" {
		days = httputil.ParseIntOrDefault(d, 30)
	}

	usages, err := h.apiUsageLogStore.GetDailyUsage(userID, days)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	resp := UsageStatsResponse{
		Message: "Daily usage stats fetched successfully",
		Data:    usages,
	}
	httputil.WriteJSON(w, http.StatusOK, resp)
}
