package handler

import (
	"megome/internal/domain/completion"
	"megome/internal/domain/user"
	"megome/internal/middleware"
	"megome/internal/pkg/httputil"
	"net/http"

	"github.com/gorilla/mux"
)

type CompletionHandler struct {
	completionStore *completion.Repository
	userStore       *user.Repository
}

func NewCompletionHandler(completionStore *completion.Repository, userStore *user.Repository) *CompletionHandler {
	return &CompletionHandler{completionStore: completionStore, userStore: userStore}
}

func (h *CompletionHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/completion", middleware.WithJWTAuth(h.handleCompletion, h.userStore)).Methods("GET")
}

type CompletionResponse struct {
	Message string                   `json:"message"`
	Data    *completion.CompletionResult `json:"data"`
}

func (h *CompletionHandler) handleCompletion(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r.Context())

	result, err := h.completionStore.GetCompletion(userID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	resp := CompletionResponse{
		Message: "Completion status fetched successfully",
		Data:    result,
	}

	httputil.WriteJSON(w, http.StatusOK, resp)
}
