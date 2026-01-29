package technology

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
	technologyStore types.TechnologyStore
	userStore       types.UserStore
}

type TechnologyResponses struct {
	Message string             `json:"message"`
	Data    []types.Technology `json:"data"`
}

func NewHandler(technologyStore types.TechnologyStore, userStore types.UserStore) *Handler {
	return &Handler{technologyStore: technologyStore, userStore: userStore}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/technology", auth.WithJWTAuth(h.handleViewTechnology, h.userStore)).Methods("GET")
	router.HandleFunc("/technology", auth.WithJWTAuth(h.handleCreateTechnology, h.userStore)).Methods("POST")
	router.HandleFunc("/technology/{id}", auth.WithJWTAuth(h.handleUpdateTechnology, h.userStore)).Methods("PUT")
	router.HandleFunc("/technology/{id}", auth.WithJWTAuth(h.handleDeleteTechnology, h.userStore)).Methods("DELETE")
}

func (h *Handler) handleViewTechnology(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserIDFromContext(r.Context())
	technology, err := h.technologyStore.GetTechnologies(userID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}
	resp := TechnologyResponses{
		Message: "Technology fetched successfully",
		Data:    technology,
	}
	utils.WriteJSON(w, http.StatusOK, resp)
}

func (h *Handler) handleCreateTechnology(w http.ResponseWriter, r *http.Request) {
	// get JSON payload
	var payload types.TechnologyPayload
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

	// create technology
	userID := auth.GetUserIDFromContext(r.Context())
	err := h.technologyStore.CreateTechnology(types.Technology{
		UserID: userID,
		Name:   payload.Name,
		Slug:   payload.Slug,
	})
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}
	utils.WriteJSON(w, http.StatusOK, map[string]string{
		"message": "Technology is successfully created",
	})
}

func (h *Handler) handleUpdateTechnology(w http.ResponseWriter, r *http.Request) {
	// get JSON payload
	var payload types.TechnologyPayload
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

	// update technology
	id, err := utils.GetRequestId(r)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}
	err = h.technologyStore.UpdateTechnology(id, types.Technology{
		Name: payload.Name,
		Slug: payload.Slug,
	})
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}
	utils.WriteJSON(w, http.StatusOK, map[string]string{"message": "Technology is successfully updated"})
}

func (h *Handler) handleDeleteTechnology(w http.ResponseWriter, r *http.Request) {
	id, err := utils.GetRequestId(r)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}
	err = h.technologyStore.DeleteTechnology(id)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}
	utils.WriteJSON(w, http.StatusOK, map[string]string{"message": "Technology is successfully deleted"})
}
