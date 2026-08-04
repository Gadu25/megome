package handler

import (
	"fmt"
	"net/http"

	"megome/internal/domain/user"
	"megome/internal/middleware"
	"megome/internal/pkg/auth"
	"megome/internal/pkg/httputil"
	"megome/internal/pkg/validator"

	playvalidator "github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
)

type AccountHandler struct {
	userStore *user.Repository
}

func NewAccountHandler(userStore *user.Repository) *AccountHandler {
	return &AccountHandler{userStore: userStore}
}

func (h *AccountHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/auth/change-email", middleware.WithJWTAuth(h.handleChangeEmail, h.userStore)).Methods("POST")
	router.HandleFunc("/auth/change-username", middleware.WithJWTAuth(h.handleChangeUsername, h.userStore)).Methods("POST")
	router.HandleFunc("/auth/account", middleware.WithJWTAuth(h.handleDeleteAccount, h.userStore)).Methods("DELETE")
}

func (h *AccountHandler) handleChangeEmail(w http.ResponseWriter, r *http.Request) {
	var payload user.ChangeEmailPayload
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

	existing, err := h.userStore.GetUserByEmail(payload.Email)
	if err == nil && existing.ID != userID {
		httputil.WriteError(w, http.StatusConflict, fmt.Errorf("email already in use"))
		return
	}

	if err := h.userStore.UpdateEmail(userID, payload.Email); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	httputil.WriteJSON(w, http.StatusOK, map[string]string{"message": "email updated successfully"})
}

func (h *AccountHandler) handleChangeUsername(w http.ResponseWriter, r *http.Request) {
	var payload user.ChangeUsernamePayload
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

	existing, err := h.userStore.GetUserByEmailOrUsername(payload.Username)
	if err == nil && existing.ID != userID {
		httputil.WriteError(w, http.StatusConflict, fmt.Errorf("username already taken"))
		return
	}

	if err := h.userStore.UpdateUsername(userID, payload.Username); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	httputil.WriteJSON(w, http.StatusOK, map[string]string{"message": "username updated successfully"})
}

func (h *AccountHandler) handleDeleteAccount(w http.ResponseWriter, r *http.Request) {
	var payload user.DeleteAccountPayload
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

	if !auth.ComparePasswords(u.Password, []byte(payload.Password)) {
		httputil.WriteError(w, http.StatusUnauthorized, fmt.Errorf("password is incorrect"))
		return
	}

	if err := h.userStore.DeleteAccount(userID); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	httputil.WriteJSON(w, http.StatusOK, map[string]string{"message": "account deleted successfully"})
}
