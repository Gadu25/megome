package api

import (
	"database/sql"
	"log"
	"megome/internal/services/certification"
	"megome/internal/services/education"
	"megome/internal/services/experience"
	"megome/internal/services/profile"
	"megome/internal/services/project"
	projecttech "megome/internal/services/projectTech"
	"megome/internal/services/refreshToken"
	"megome/internal/services/skill"
	"megome/internal/services/technology"
	"megome/internal/services/user"
	"net/http"

	"github.com/gorilla/mux"
)

type APIServer struct {
	addr string
	db   *sql.DB
}

func NewAPIServer(addr string, db *sql.DB) *APIServer {
	return &APIServer{
		addr: addr,
		db:   db,
	}
}

func (s *APIServer) Run() error {
	router := mux.NewRouter()
	subrouter := router.PathPrefix("/api/v1").Subrouter()

	refreshStore := refreshToken.NewStore(s.db)

	userStore := user.NewStore(s.db)
	userHandler := user.NewHandler(userStore, refreshStore)
	userHandler.RegisterRoutes(subrouter)

	profileStore := profile.NewStore(s.db)
	profileHandler := profile.NewHandler(profileStore, userStore)
	profileHandler.RegisterRoutes(subrouter)

	experienceStore := experience.NewStore(s.db)
	experienceHandler := experience.NewHandler(experienceStore, userStore)
	experienceHandler.RegisterRoutes(subrouter)

	skillStore := skill.NewStore(s.db)
	skillHandler := skill.NewHandler(skillStore, userStore)
	skillHandler.RegisterRoutes(subrouter)

	educationStore := education.NewStore(s.db)
	educationHandler := education.NewHandler(educationStore, userStore)
	educationHandler.RegisterRoutes(subrouter)

	certificationStore := certification.NewStore(s.db)
	certificationHandler := certification.NewHandler(certificationStore, userStore)
	certificationHandler.RegisterRoutes(subrouter)

	technologyStore := technology.NewStore(s.db)
	technologyHandler := technology.NewHandler(technologyStore, userStore)
	technologyHandler.RegisterRoutes(subrouter)

	projectStore := project.NewStore(s.db)
	projectHandler := project.NewHandler(projectStore, userStore)
	projectHandler.RegisterRoutes(subrouter)

	projectTechStore := projecttech.NewStore(s.db)
	projectTechHandler := projecttech.NewHandler(projectTechStore, userStore)
	projectTechHandler.RegisterRoutes(subrouter)

	log.Println("Listening on", s.addr)
	return http.ListenAndServe(s.addr, router)
}
