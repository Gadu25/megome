package handler

import (
	"fmt"
	"net/http"

	"megome/internal/domain/refreshtoken"
	"megome/internal/domain/user"
	"megome/internal/middleware"
	"megome/internal/pkg/auth"
	"megome/internal/pkg/httputil"
	"megome/internal/pkg/validator"

	playvalidator "github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
)

type SecurityHandler struct {
	userStore    *user.Repository
	refreshStore *refreshtoken.Repository
}

func NewSecurityHandler(userStore *user.Repository, refreshStore *refreshtoken.Repository) *SecurityHandler {
	return &SecurityHandler{userStore: userStore, refreshStore: refreshStore}
}

func (h *SecurityHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/auth/change-password", middleware.WithJWTAuth(h.handleChangePassword, h.userStore)).Methods("POST")
	router.HandleFunc("/auth/sessions", middleware.WithJWTAuth(h.handleSessions, h.userStore)).Methods("GET")
	router.HandleFunc("/auth/sessions/{id}/revoke", middleware.WithJWTAuth(h.handleRevokeSession, h.userStore)).Methods("POST")
}

func (h *SecurityHandler) handleChangePassword(w http.ResponseWriter, r *http.Request) {
	var payload user.ChangePasswordPayload
	if err := httputil.ParseJSON(r, &payload); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err)
		return
	}
	if err := validator.Validate.Struct(payload); err != nil {
		errors := err.(playvalidator.ValidationErrors)
		httputil.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid payload %v", errors))
		return
	}

	userID := middleware.GetUserIDFromContext(r.Context())
	u, err := h.userStore.GetUserByID(userID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	if !auth.ComparePasswords(u.Password, []byte(payload.CurrentPassword)) {
		httputil.WriteError(w, http.StatusUnauthorized, fmt.Errorf("current password is incorrect"))
		return
	}

	hashedPassword, err := auth.HashedPassword(payload.NewPassword)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	if err := h.userStore.UpdatePassword(userID, hashedPassword); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	httputil.WriteJSON(w, http.StatusOK, map[string]string{"message": "password changed successfully"})
}

func (h *SecurityHandler) handleSessions(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r.Context())

	sessions, err := h.refreshStore.ListActiveSessions(userID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	if sessions == nil {
		sessions = []refreshtoken.SessionInfo{}
	}

	httputil.WriteJSON(w, http.StatusOK, map[string]interface{}{"sessions": sessions})
}

func (h *SecurityHandler) handleRevokeSession(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r.Context())

	sessionID, err := httputil.GetRequestId(r)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid session id"))
		return
	}

	if err := h.refreshStore.RevokeSession(sessionID, userID); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	httputil.WriteJSON(w, http.StatusOK, map[string]string{"message": "session revoked"})
}
