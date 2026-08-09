package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"megome/internal/domain/certification"
	"megome/internal/domain/education"
	"megome/internal/domain/experience"
	"megome/internal/domain/profile"
	"megome/internal/domain/project"
	"megome/internal/domain/skill"
	"megome/internal/domain/user"
	"megome/internal/middleware"
	"megome/internal/pkg/httputil"

	"github.com/gorilla/mux"
)

type DataExportHandler struct {
	userStore          *user.Repository
	profileStore       *profile.Repository
	skillStore         *skill.Repository
	educationStore     *education.Repository
	experienceStore    *experience.Repository
	projectStore       *project.Repository
	certificationStore *certification.Repository
}

func NewDataExportHandler(
	userStore *user.Repository,
	profileStore *profile.Repository,
	skillStore *skill.Repository,
	educationStore *education.Repository,
	experienceStore *experience.Repository,
	projectStore *project.Repository,
	certificationStore *certification.Repository,
) *DataExportHandler {
	return &DataExportHandler{
		userStore:          userStore,
		profileStore:       profileStore,
		skillStore:         skillStore,
		educationStore:     educationStore,
		experienceStore:    experienceStore,
		projectStore:       projectStore,
		certificationStore: certificationStore,
	}
}

type ExportData struct {
	ExportedAt     string                       `json:"exportedAt"`
	Profile        any                          `json:"profile"`
	Experiences    []experience.Experience      `json:"experiences"`
	Skills         []skill.Skill                `json:"skills"`
	Education      []education.Education        `json:"education"`
	Projects       []project.Project            `json:"projects"`
	Certifications []certification.Certification `json:"certifications"`
}

func (h *DataExportHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/data/export", middleware.WithJWTAuth(h.handleExport, h.userStore)).Methods("GET")
}

func (h *DataExportHandler) handleExport(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r.Context())

	profile, err := h.profileStore.GetProfile(userID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	skills, err := h.skillStore.GetSkills(userID, 1000, 0)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	experiences, err := h.experienceStore.GetExperiences(userID, 1000, 0)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	education, err := h.educationStore.GetEducations(userID, 1000, 0)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	projects, err := h.projectStore.GetProjects(userID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	certifications, err := h.certificationStore.GetCertifications(userID, 1000, 0)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	export := ExportData{
		ExportedAt:     time.Now().UTC().Format(time.RFC3339),
		Profile:        profile,
		Experiences:    experiences,
		Skills:         skills,
		Education:      education,
		Projects:       projects,
		Certifications: certifications,
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", "attachment; filename=megome-export.json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(export)
}
