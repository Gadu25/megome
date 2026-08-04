package handler

import (
	"errors"
	"net/http"

	"megome/internal/domain/assist"
	"megome/internal/domain/user"
	"megome/internal/middleware"
	"megome/internal/pkg/ai"
	"megome/internal/pkg/httputil"

	"github.com/gorilla/mux"
)

type AssistHandler struct {
	aiService *ai.Service
	userStore *user.Repository
}

func NewAssistHandler(aiService *ai.Service, userStore *user.Repository) *AssistHandler {
	return &AssistHandler{aiService: aiService, userStore: userStore}
}

func (h *AssistHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/ai/assist", middleware.WithJWTAuth(h.handleAssist, h.userStore)).Methods("POST")
	router.HandleFunc("/ai/status", middleware.WithJWTAuth(h.handleStatus, h.userStore)).Methods("GET")
}

func (h *AssistHandler) handleAssist(w http.ResponseWriter, r *http.Request) {
	var req assist.Request
	if err := httputil.ParseJSON(r, &req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err)
		return
	}

	fields, err := h.aiService.Assist(r.Context(), req.Task, req.Context, req.Extra)
	if err != nil {
		var ua *ai.UnavailableError
		if errors.As(err, &ua) {
			httputil.WriteJSON(w, http.StatusTooManyRequests, map[string]any{
				"message": "ai_unavailable",
				"data": map[string]any{
					"cooldownRemainingSeconds": ua.RemainingSeconds,
				},
			})
			return
		}
		if errors.Is(err, ai.ErrUnknownTask) {
			httputil.WriteError(w, http.StatusBadRequest, err)
			return
		}
		httputil.WriteError(w, http.StatusBadGateway, err)
		return
	}

	httputil.WriteJSON(w, http.StatusOK, map[string]any{
		"message": "AI content generated successfully",
		"data": map[string]any{
			"task":   req.Task,
			"fields": fields,
		},
	})
}

func (h *AssistHandler) handleStatus(w http.ResponseWriter, r *http.Request) {
	available, remaining := h.aiService.Status()
	httputil.WriteJSON(w, http.StatusOK, map[string]any{
		"message": "AI status fetched successfully",
		"data": map[string]any{
			"available":                available,
			"cooldownRemainingSeconds": remaining,
		},
	})
}
