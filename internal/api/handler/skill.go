package handler

import (
	"fmt"
	"megome/internal/domain/skill"
	"megome/internal/domain/user"
	"megome/internal/middleware"
	"megome/internal/pkg/httputil"
	"megome/internal/pkg/validator"
	"net/http"

	playvalidator "github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
)

type SkillHandler struct {
	skillStore *skill.Repository
	userStore  *user.Repository
}

type SkillReponse struct {
	Message string        `json:"message"`
	Skills  []skill.Skill `json:"skills"`
}

type SingleSkillResponse struct {
	Message string      `json:"message"`
	Skill   skill.Skill `json:"skill"`
}

func NewSkillHandler(skillStore *skill.Repository, userStore *user.Repository) *SkillHandler {
	return &SkillHandler{skillStore: skillStore, userStore: userStore}
}

func (h *SkillHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/skill", middleware.WithJWTAuth(h.handleViewSkills, h.userStore)).Methods("GET")
	router.HandleFunc("/skill", middleware.WithJWTAuth(h.handleCreateSkill, h.userStore)).Methods("POST")
	router.HandleFunc("/skill/{id}", middleware.WithJWTAuth(h.handleUpdateSkill, h.userStore)).Methods("PUT")
	router.HandleFunc("/skill/{id}", middleware.WithJWTAuth(h.handleDeleteSkill, h.userStore)).Methods("DELETE")
}

func (h *SkillHandler) handleViewSkills(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r.Context())
	skills, err := h.skillStore.GetSkills(userID)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err)
		return
	}
	resp := SkillReponse{
		Message: "Skills fetched successfully",
		Skills:  skills,
	}
	httputil.WriteJSON(w, http.StatusOK, resp)
}

func (h *SkillHandler) handleCreateSkill(w http.ResponseWriter, r *http.Request) {
	var payload skill.SkillPayload
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
	s, err := h.skillStore.CreateSkill(skill.Skill{
		UserID:      userID,
		SkillName:   payload.SkillName,
		Proficiency: payload.Proficiency,
	})

	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	resp := SingleSkillResponse{
		Message: "Skill added successfully",
		Skill:   s,
	}
	httputil.WriteJSON(w, http.StatusOK, resp)
}

func (h *SkillHandler) handleUpdateSkill(w http.ResponseWriter, r *http.Request) {
	var payload skill.SkillPayload
	if err := httputil.ParseJSON(r, &payload); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err)
		return
	}

	if err := validator.Validate.Struct(payload); err != nil {
		errors := err.(playvalidator.ValidationErrors)
		httputil.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid payload %v", errors))
		return
	}

	id, err := httputil.GetRequestId(r)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err)
		return
	}

	s, err := h.skillStore.UpdateSkill(id, skill.Skill{
		SkillName:   payload.SkillName,
		Proficiency: payload.Proficiency,
	})

	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	resp := SingleSkillResponse{
		Message: "Skill updated successfully",
		Skill:   s,
	}
	httputil.WriteJSON(w, http.StatusOK, resp)
}

func (h *SkillHandler) handleDeleteSkill(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.GetRequestId(r)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err)
		return
	}
	s, err := h.skillStore.DeleteSkill(id)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	resp := SingleSkillResponse{
		Message: "Skill deleted successfully",
		Skill:   s,
	}
	httputil.WriteJSON(w, http.StatusOK, resp)
}
