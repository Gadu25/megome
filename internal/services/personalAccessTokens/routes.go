package personalaccesstokens

import (
	"fmt"
	"megome/internal/services/auth"
	"megome/internal/services/types"
	"megome/internal/services/utils"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
)

type Handler struct {
	userStore types.UserStore
	patStore  types.PersonalAccessTokenStore
}

type PATListResponse struct {
	Message string                      `json:"message"`
	PATs    []types.PersonalAccessToken `json:"pats"`
}

type PATResponse struct {
	Message string `json:"message"`
	PAT     string `json:"pat"`
}

type PATCountResponse struct {
	Message      string `json:"mesage"`
	UserPatCount int    `json:"userPatCount"`
}

func NewHandler(userStore types.UserStore, patStore types.PersonalAccessTokenStore) *Handler {
	return &Handler{userStore: userStore, patStore: patStore}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/pat", auth.WithJWTAuth(h.handleViewPATs, h.userStore)).Methods("GET")
	router.HandleFunc("/pat", auth.WithJWTAuth(h.handleCreatePAT, h.userStore)).Methods("POST")
	router.HandleFunc("/pat/{id}/revoke", auth.WithJWTAuth(h.handleRevokePAT, h.userStore)).Methods("POST")
	router.HandleFunc("/pat/{id}", auth.WithJWTAuth(h.handleDeletePAT, h.userStore)).Methods("DELETE")
	router.HandleFunc("/pat/count", auth.WithJWTAuth(h.handleViewUserPATCount, h.userStore)).Methods("GET")
}

func (h *Handler) handleViewPATs(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserIDFromContext(r.Context())

	pats, err := h.patStore.GetPATs(userID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	resp := PATListResponse{
		Message: "Personal access tokens fetched successfully",
		PATs:    pats,
	}

	utils.WriteJSON(w, http.StatusOK, resp)
}

func (h *Handler) handleCreatePAT(w http.ResponseWriter, r *http.Request) {
	// get JSON payload
	var payload types.PersonalAccessTokenPayload
	if err := utils.ParseJSON(r, &payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	// validate the payload
	if err := utils.Validate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid payload %v", errors))
		return
	}

	userID := auth.GetUserIDFromContext(r.Context())
	pat, err := h.patStore.CreatePAT(userID, payload.Name)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	resp := PATResponse{
		Message: "Token is Successfully created!",
		PAT:     pat,
	}

	utils.WriteJSON(w, http.StatusOK, resp)
}

func (h *Handler) handleRevokePAT(w http.ResponseWriter, r *http.Request) {

	userID := auth.GetUserIDFromContext(r.Context())

	id, err := utils.GetRequestId(r)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	err = h.patStore.RevokePAT(userID, id)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]string{
		"message": "token revoked successfully",
	})
}

func (h *Handler) handleDeletePAT(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserIDFromContext(r.Context())

	id, err := utils.GetRequestId(r)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	err = h.patStore.DeletePAT(userID, id)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]string{
		"message": "token deleted successfully",
	})
}

func (h *Handler) handleViewUserPATCount(w http.ResponseWriter, r *http.Request) {
	userId := auth.GetPATTokenIDFromContext(r.Context())
	count, err := h.patStore.GetTokenCountByUserID(userId)

	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	resp := PATCountResponse{
		Message:      "User personal access token count",
		UserPatCount: count,
	}

	utils.WriteJSON(w, http.StatusOK, resp)
}
