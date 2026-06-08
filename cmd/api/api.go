package api

import (
	"database/sql"
	"log"
	"megome/config"
	"megome/internal/platform/http/middleware"
	apilogs "megome/internal/services/apiLogs"
	"megome/internal/services/certification"
	"megome/internal/services/education"
	"megome/internal/services/experience"
	"megome/internal/services/initData"
	personalaccesstokens "megome/internal/services/personalAccessTokens"
	"megome/internal/services/profile"
	"megome/internal/services/project"
	projectimages "megome/internal/services/projectImages"
	projecttech "megome/internal/services/projectTech"
	publiccertification "megome/internal/services/public/certification"
	publiceducation "megome/internal/services/public/education"
	publicexperience "megome/internal/services/public/experience"
	publicprofile "megome/internal/services/public/profile"
	publicproject "megome/internal/services/public/project"
	publicskill "megome/internal/services/public/skill"
	"megome/internal/services/refreshToken"
	"megome/internal/services/skill"
	"megome/internal/services/storage"
	"megome/internal/services/technology"
	"megome/internal/services/user"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

type APIServer struct {
	addr string
	db   *sql.DB
}

func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		path := r.URL.Path

		// INTERNAL API (strict)
		if strings.HasPrefix(path, "/api/v1") {
			w.Header().Set("Access-Control-Allow-Origin", config.Envs.FrontendUrl)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		// PUBLIC API (open or developer-friendly)
		if strings.HasPrefix(path, "/public/v1") {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
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
	internal := router.PathPrefix("/api/v1").Subrouter()
	public := router.PathPrefix("/public/v1").Subrouter()

	// API rate limiter
	rateLimiter := middleware.NewRateLimiter(4, 10)

	// Apply middleware to route groups
	internal.Use(rateLimiter.Middleware)
	public.Use(rateLimiter.Middleware)

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
	refreshHandler.RegisterRoutes(internal)

	profileStore := profile.NewStore(s.db)

	userStore := user.NewStore(s.db)
	userHandler := user.NewHandler(userStore, profileStore, refreshStore)
	userHandler.RegisterRoutes(internal)

	profileHandler := profile.NewHandler(profileStore, userStore, r2Client)
	profileHandler.RegisterRoutes(internal)

	initHandler := initData.NewHandler(profileStore, userStore)
	initHandler.RegisterRoutes(internal)

	experienceStore := experience.NewStore(s.db)
	experienceHandler := experience.NewHandler(experienceStore, userStore)
	experienceHandler.RegisterRoutes(internal)

	skillStore := skill.NewStore(s.db)
	skillHandler := skill.NewHandler(skillStore, userStore)
	skillHandler.RegisterRoutes(internal)

	educationStore := education.NewStore(s.db)
	educationHandler := education.NewHandler(educationStore, userStore)
	educationHandler.RegisterRoutes(internal)

	certificationStore := certification.NewStore(s.db)
	certificationHandler := certification.NewHandler(certificationStore, userStore)
	certificationHandler.RegisterRoutes(internal)

	technologyStore := technology.NewStore(s.db)
	technologyHandler := technology.NewHandler(technologyStore, userStore)
	technologyHandler.RegisterRoutes(internal)

	projectStore := project.NewStore(s.db, r2Client)
	projectHandler := project.NewHandler(projectStore, userStore)
	projectHandler.RegisterRoutes(internal)

	projectImageStore := projectimages.NewStore(s.db)
	projectImageHandler := projectimages.NewHandler(projectImageStore, userStore, r2Client)
	projectImageHandler.RegisterRoutes(internal)

	projectTechStore := projecttech.NewStore(s.db)
	projectTechHandler := projecttech.NewHandler(projectTechStore, userStore)
	projectTechHandler.RegisterRoutes(internal)

	personalAccessTokenStore := personalaccesstokens.NewStore(s.db)
	personalAccesstokenHandler := personalaccesstokens.NewHandler(userStore, personalAccessTokenStore)
	personalAccesstokenHandler.RegisterRoutes(internal)

	apiUsageLog := apilogs.NewStore(s.db)
	apiUsagLogHandler := apilogs.NewHandler(apiUsageLog, userStore)
	apiUsagLogHandler.RegisterRoutes(internal)

	// PUBLIC
	apiLogStore := apilogs.NewStore(s.db)

	publicProfileHandler := publicprofile.NewHandler(profileStore, personalAccessTokenStore, apiLogStore)
	publicProfileHandler.RegisterRoutes(public)

	publicSkillHandler := publicskill.NewHandler(skillStore, personalAccessTokenStore, apiLogStore)
	publicSkillHandler.RegisterRoutes(public)

	publicEducationHandler := publiceducation.NewHandler(educationStore, personalAccessTokenStore, apiLogStore)
	publicEducationHandler.RegisterRoutes(public)

	publicProjectHandler := publicproject.NewHandler(projectStore, personalAccessTokenStore, apiLogStore)
	publicProjectHandler.RegisterRoutes(public)

	publicExperienceHandler := publicexperience.NewHandler(experienceStore, personalAccessTokenStore, apiLogStore)
	publicExperienceHandler.RegisterRoutes(public)

	publicCertificateHandler := publiccertification.NewHandler(certificationStore, personalAccessTokenStore, apiLogStore)
	publicCertificateHandler.RegisterRoutes(public)

	// for CORS
	corsRouter := CORS(router)

	log.Println("Listening on", s.addr)
	return http.ListenAndServe(s.addr, corsRouter)
}
