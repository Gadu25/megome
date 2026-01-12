package api

import (
	"database/sql"
	"log"
	"megome/internal/services/education"
	"megome/internal/services/experience"
	"megome/internal/services/profile"
	"megome/internal/services/skill"
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

	userStore := user.NewStore(s.db)
	userHandler := user.NewHandler(userStore)
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

	log.Println("Listening on", s.addr)
	return http.ListenAndServe(s.addr, router)
}
