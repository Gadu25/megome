package handler

import (
	"megome/internal/domain/apilog"
	"megome/internal/domain/user"
	"megome/internal/middleware"
	"megome/internal/pkg/httputil"
	"net/http"

	"github.com/gorilla/mux"
)

type APILogHandler struct {
	apiUsageLogStore *apilog.Repository
	userStore        *user.Repository
}

type APILogResponse struct {
	Message string                    `json:"message"`
	Data    apilog.APIUsageLogWithToken `json:"data"`
}

type UsageResponse struct {
	Message string                   `json:"message"`
	Data    apilog.UserAPIUsageStats `json:"data"`
}

func NewAPILogHandler(apiUsageLogStore *apilog.Repository, userStore *user.Repository) *APILogHandler {
	return &APILogHandler{apiUsageLogStore: apiUsageLogStore, userStore: userStore}
}

func (h *APILogHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/api-logs/token/{id}", middleware.WithJWTAuth(h.handleViewLog, h.userStore)).Methods("GET")
	router.HandleFunc("/api-logs/usage", middleware.WithJWTAuth(h.handleViewUserUsage, h.userStore)).Methods("GET")
}

func (h *APILogHandler) handleViewLog(w http.ResponseWriter, r *http.Request) {
	tokenID, err := httputil.GetRequestId(r)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err)
		return
	}

	limit := 50
	offset := 0

	query := r.URL.Query()

	if l := query.Get("limit"); l != "" {
		limit = httputil.ParseIntOrDefault(l, 50)
	}

	if o := query.Get("offset"); o != "" {
		offset = httputil.ParseIntOrDefault(o, 0)
	}

	data, err := h.apiUsageLogStore.GetByTokenID(tokenID, limit, offset)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	resp := APILogResponse{
		Message: "API usage logs fetched successfully",
		Data:    data,
	}

	httputil.WriteJSON(w, http.StatusOK, resp)
}

func (h *APILogHandler) handleViewUserUsage(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetPATTokenIDFromContext(r.Context())
	stats, err := h.apiUsageLogStore.GetUserUsageStats(userID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err)
		return
	}
	resp := UsageResponse{
		Message: "User public api usage",
		Data:    stats,
	}
	httputil.WriteJSON(w, http.StatusOK, resp)
}
