package handler

import (
	"fmt"
	"megome/internal/domain/personalaccesstoken"
	"megome/internal/domain/user"
	"megome/internal/middleware"
	"megome/internal/pkg/httputil"
	"megome/internal/pkg/validator"
	"net/http"

	playvalidator "github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
)

type PersonalAccessTokenHandler struct {
	userStore *user.Repository
	patStore  *personalaccesstoken.Repository
}

type PATListResponse struct {
	Message string                          `json:"message"`
	PATs    []personalaccesstoken.PersonalAccessToken `json:"pats"`
}

type PATResponse struct {
	Message string `json:"message"`
	PAT     string `json:"pat"`
}

type PATCountResponse struct {
	Message      string `json:"message"`
	UserPatCount int    `json:"userPatCount"`
}

func NewPersonalAccessTokenHandler(userStore *user.Repository, patStore *personalaccesstoken.Repository) *PersonalAccessTokenHandler {
	return &PersonalAccessTokenHandler{userStore: userStore, patStore: patStore}
}

func (h *PersonalAccessTokenHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/pat", middleware.WithJWTAuth(h.handleViewPATs, h.userStore)).Methods("GET")
	router.HandleFunc("/pat", middleware.WithJWTAuth(h.handleCreatePAT, h.userStore)).Methods("POST")
	router.HandleFunc("/pat/{id}/revoke", middleware.WithJWTAuth(h.handleRevokePAT, h.userStore)).Methods("POST")
	router.HandleFunc("/pat/{id}", middleware.WithJWTAuth(h.handleDeletePAT, h.userStore)).Methods("DELETE")
	router.HandleFunc("/pat/count", middleware.WithJWTAuth(h.handleViewUserPATCount, h.userStore)).Methods("GET")
}

func (h *PersonalAccessTokenHandler) handleViewPATs(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r.Context())

	limit := 20
	offset := 0
	query := r.URL.Query()
	if l := query.Get("limit"); l != "" {
		limit = httputil.ParseIntOrDefault(l, 20)
	}
	if o := query.Get("offset"); o != "" {
		offset = httputil.ParseIntOrDefault(o, 0)
	}
	if limit > 100 {
		limit = 100
	}

	pats, err := h.patStore.GetPATs(userID, limit, offset)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	total, err := h.patStore.GetTokenCountByUserID(userID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	resp := personalaccesstoken.PaginatedPATResponse{
		Data: pats,
	}
	resp.Pagination.Limit = limit
	resp.Pagination.Offset = offset
	resp.Pagination.Total = total
	httputil.WriteJSON(w, http.StatusOK, resp)
}

func (h *PersonalAccessTokenHandler) handleCreatePAT(w http.ResponseWriter, r *http.Request) {
	var payload personalaccesstoken.PersonalAccessTokenPayload
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
	pat, err := h.patStore.CreatePAT(userID, payload.Name)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	resp := PATResponse{
		Message: "Token successfully created",
		PAT:     pat,
	}

	httputil.WriteJSON(w, http.StatusOK, resp)
}

func (h *PersonalAccessTokenHandler) handleRevokePAT(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r.Context())

	id, err := httputil.GetRequestId(r)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err)
		return
	}

	err = h.patStore.RevokePAT(userID, id)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err)
		return
	}

	httputil.WriteJSON(w, http.StatusOK, map[string]string{
		"message": "token revoked successfully",
	})
}

func (h *PersonalAccessTokenHandler) handleDeletePAT(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r.Context())

	id, err := httputil.GetRequestId(r)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err)
		return
	}

	err = h.patStore.DeletePAT(userID, id)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err)
		return
	}

	httputil.WriteJSON(w, http.StatusOK, map[string]string{
		"message": "token deleted successfully",
	})
}

func (h *PersonalAccessTokenHandler) handleViewUserPATCount(w http.ResponseWriter, r *http.Request) {
	userId := middleware.GetPATTokenIDFromContext(r.Context())
	count, err := h.patStore.GetTokenCountByUserID(userId)

	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err)
		return
	}

	resp := PATCountResponse{
		Message:      "User personal access token count",
		UserPatCount: count,
	}

	httputil.WriteJSON(w, http.StatusOK, resp)
}
