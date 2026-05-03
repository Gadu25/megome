package api

import (
	"database/sql"
	"log"
	"megome/config"
	"megome/internal/services/certification"
	"megome/internal/services/education"
	"megome/internal/services/experience"
	"megome/internal/services/initData"
	"megome/internal/services/profile"
	"megome/internal/services/project"
	projectimages "megome/internal/services/projectImages"
	projecttech "megome/internal/services/projectTech"
	"megome/internal/services/refreshToken"
	"megome/internal/services/skill"
	"megome/internal/services/storage"
	"megome/internal/services/technology"
	"megome/internal/services/user"
	"net/http"

	"github.com/gorilla/mux"
)

type APIServer struct {
	addr string
	db   *sql.DB
}

func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Allow frontend origin
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
		w.Header().Set("Access-Control-Allow-Credentials", "true") // important for cookies
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Handle preflight request
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
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

	r2Cfg := storage.Config{
		AccessKey: config.Envs.R2AccessKeyId,
		SecretKey: config.Envs.R2SecretAccessKey,
		Bucket:    config.Envs.R2Bucket,
		Endpoint:  config.Envs.R2Endpoint,
	}

	r2Client, err := storage.NewR2Client(r2Cfg)
	if err != nil {
		log.Fatalf("failed to initialize R2 client: %v", err)
	}

	refreshStore := refreshToken.NewStore(s.db)
	refreshHandler := refreshToken.NewHandler(refreshStore)
	refreshHandler.RegisterRoutes(subrouter)

	userStore := user.NewStore(s.db)
	userHandler := user.NewHandler(userStore, refreshStore)
	userHandler.RegisterRoutes(subrouter)

	profileStore := profile.NewStore(s.db)
	profileHandler := profile.NewHandler(profileStore, userStore, r2Client)
	profileHandler.RegisterRoutes(subrouter)

	initHandler := initData.NewHandler(profileStore, userStore)
	initHandler.RegisterRoutes(subrouter)

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

	projectImageStore := projectimages.NewStore(s.db)
	projectImageHandler := projectimages.NewHandler(projectImageStore, userStore, r2Client)
	projectImageHandler.RegisterRoutes(subrouter)

	projectTechStore := projecttech.NewStore(s.db)
	projectTechHandler := projecttech.NewHandler(projectTechStore, userStore)
	projectTechHandler.RegisterRoutes(subrouter)

	// for CORS
	corsRouter := CORS(router)

	log.Println("Listening on", s.addr)
	return http.ListenAndServe(s.addr, corsRouter)
}
