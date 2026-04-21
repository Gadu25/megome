package skill

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
	skillStore types.SkillStore
	userStore  types.UserStore
}

type SkillReponse struct {
	Message string        `json:"message"`
	Skills  []types.Skill `json:"skills"`
}

func NewHandler(skillStore types.SkillStore, userStore types.UserStore) *Handler {
	return &Handler{skillStore: skillStore, userStore: userStore}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/skill", auth.WithJWTAuth(h.handleViewSkills, h.userStore)).Methods("GET")
	router.HandleFunc("/skill", auth.WithJWTAuth(h.handleCreateSkill, h.userStore)).Methods("POST")
	router.HandleFunc("/skill/{id}", auth.WithJWTAuth(h.handleUpdateSkill, h.userStore)).Methods("PUT")
	router.HandleFunc("/skill/{id}", auth.WithJWTAuth(h.handleDeleteSkill, h.userStore)).Methods("DELETE")
}

func (h *Handler) handleViewSkills(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserIDFromContext(r.Context())
	skills, err := h.skillStore.GetSkills(userID)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}
	resp := SkillReponse{
		Message: "Skills fetched successfully",
		Skills:  skills,
	}
	utils.WriteJSON(w, http.StatusOK, resp)
}

func (h *Handler) handleCreateSkill(w http.ResponseWriter, r *http.Request) {
	// get JSON payload
	payload := types.SkillPayload{
		SkillName:   r.FormValue("skillName"),
		Proficiency: r.FormValue("proficiency"),
	}

	// validate the payload
	if err := utils.Validate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid payload %v", errors))
		return
	}

	// create skill
	userID := auth.GetUserIDFromContext(r.Context())
	err := h.skillStore.CreateSkill(types.Skill{
		UserID:      userID,
		SkillName:   payload.SkillName,
		Proficiency: payload.Proficiency,
	})

	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	skills, err := h.skillStore.GetSkills(userID)

	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	resp := SkillReponse{
		Message: "Skill added successfully",
		Skills:  skills,
	}
	utils.WriteJSON(w, http.StatusOK, resp)
}

func (h *Handler) handleUpdateSkill(w http.ResponseWriter, r *http.Request) {
	// get JSON payload
	payload := types.SkillPayload{
		SkillName:   r.FormValue("skillName"),
		Proficiency: r.FormValue("proficiency"),
	}

	// validate the payload
	if err := utils.Validate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid payload %v", errors))
		return
	}

	// update skill
	id, err := utils.GetRequestId(r)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}
	err = h.skillStore.UpdateSkill(id, types.Skill{
		SkillName:   payload.SkillName,
		Proficiency: payload.Proficiency,
	})
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	userID := auth.GetUserIDFromContext(r.Context())
	skills, err := h.skillStore.GetSkills(userID)

	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	resp := SkillReponse{
		Message: "Skill updated successfully",
		Skills:  skills,
	}
	utils.WriteJSON(w, http.StatusOK, resp)
}

func (h *Handler) handleDeleteSkill(w http.ResponseWriter, r *http.Request) {
	id, err := utils.GetRequestId(r)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}
	err = h.skillStore.DeleteSkill(id)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}
	utils.WriteJSON(w, http.StatusOK, map[string]string{"message": "Skill is successfully deleted"})
}
