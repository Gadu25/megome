package apilogs

import (
	"megome/internal/services/auth"
	"megome/internal/services/types"
	"megome/internal/services/utils"
	"net/http"

	"github.com/gorilla/mux"
)

type Handler struct {
	apiUsageLogStore types.APIUsageLogStore
	userStore        types.UserStore
}

type Response struct {
	Message string                     `json:"message"`
	Data    types.APIUsageLogWithToken `json:"data"`
}

func NewHandler(apiUsageLogStore types.APIUsageLogStore, userStore types.UserStore) *Handler {
	return &Handler{apiUsageLogStore: apiUsageLogStore, userStore: userStore}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/api-logs/token/{id}", auth.WithJWTAuth(h.handleViewLog, h.userStore)).Methods("GET")
}

func (h *Handler) handleViewLog(w http.ResponseWriter, r *http.Request) {
	tokenID, err := utils.GetRequestId(r)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	// optional pagination (recommended)
	limit := 50
	offset := 0

	query := r.URL.Query()

	if l := query.Get("limit"); l != "" {
		limit = utils.ParseIntOrDefault(l, 50)
	}

	if o := query.Get("offset"); o != "" {
		offset = utils.ParseIntOrDefault(o, 0)
	}

	data, err := h.apiUsageLogStore.GetByTokenID(tokenID, limit, offset)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	resp := Response{
		Message: "API usage logs fetched successfully",
		Data:    data,
	}

	utils.WriteJSON(w, http.StatusOK, resp)
}
