package api

import (
	"database/sql"
	"log"
	"net/http"
	"strings"
	"time"

	"megome/internal/api/handler"
	"megome/internal/api/handler/public"
	"megome/internal/config"
	"megome/internal/domain/apilog"
	"megome/internal/domain/certification"
	"megome/internal/domain/completion"
	"megome/internal/domain/education"
	"megome/internal/domain/experience"
	"megome/internal/domain/passwordforgot"
	"megome/internal/domain/personalaccesstoken"
	"megome/internal/domain/profile"
	"megome/internal/domain/project"
	"megome/internal/domain/refreshtoken"
	"megome/internal/domain/skill"
	"megome/internal/domain/technology"
	"megome/internal/domain/user"
	"megome/internal/middleware"
	"megome/internal/pkg/ai"
	"megome/internal/pkg/mailer"
	"megome/internal/pkg/storage"

	"github.com/gorilla/mux"
)

type APIServer struct {
	addr string
	db   *sql.DB
}

func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		if strings.HasPrefix(path, "/api/v1") {
			w.Header().Set("Access-Control-Allow-Origin", config.Envs.FrontendUrl)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}

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
	publicRouter := router.PathPrefix("/public/v1").Subrouter()

	rateLimiter := middleware.NewRateLimiter(4, 10)

	internal.Use(rateLimiter.Middleware)
	publicRouter.Use(rateLimiter.Middleware)

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

	smtpClient := mailer.New(mailer.Config{
		Host:     config.Envs.SmtpHost,
		Port:     config.Envs.SmtpPort,
		Username: config.Envs.SmtpUsername,
		Password: config.Envs.SmtpPassword,
		From:     config.Envs.SmtpFrom,
	})

	renderer, _ := mailer.NewRenderer("internal/pkg/mailer/templates/*.html")
	emailService := mailer.NewService(smtpClient, renderer)

	userRepo := user.NewRepository(s.db)
	profileRepo := profile.NewRepository(s.db)
	refreshRepo := refreshtoken.NewRepository(s.db)
	passwordForgotRepo := passwordforgot.NewRepository(s.db)
	experienceRepo := experience.NewRepository(s.db)
	skillRepo := skill.NewRepository(s.db)
	educationRepo := education.NewRepository(s.db)
	certificationRepo := certification.NewRepository(s.db)
	technologyRepo := technology.NewRepository(s.db)
	projectRepo := project.NewRepository(s.db, r2Client)
	patRepo := personalaccesstoken.NewRepository(s.db)
	apiLogRepo := apilog.NewRepository(s.db)
	completionRepo := completion.NewRepository(s.db)

	handler.NewRefreshTokenHandler(refreshRepo).RegisterRoutes(internal)
	handler.NewUserHandler(userRepo, profileRepo, refreshRepo, emailService, passwordForgotRepo).RegisterRoutes(internal)
	handler.NewProfileHandler(profileRepo, userRepo, r2Client).RegisterRoutes(internal)
	handler.NewInitDataHandler(profileRepo, userRepo).RegisterRoutes(internal)
	handler.NewExperienceHandler(experienceRepo, userRepo, r2Client).RegisterRoutes(internal)
	handler.NewSkillHandler(skillRepo, userRepo).RegisterRoutes(internal)
	handler.NewEducationHandler(educationRepo, userRepo).RegisterRoutes(internal)
	handler.NewCertificationHandler(certificationRepo, userRepo, r2Client).RegisterRoutes(internal)
	handler.NewTechnologyHandler(technologyRepo, userRepo).RegisterRoutes(internal)
	handler.NewProjectHandler(projectRepo, userRepo).RegisterRoutes(internal)
	handler.NewProjectImageHandler(projectRepo, userRepo, r2Client).RegisterRoutes(internal)
	handler.NewProjectTechHandler(projectRepo, userRepo).RegisterRoutes(internal)
	handler.NewExperienceTechHandler(experienceRepo, userRepo).RegisterRoutes(internal)
	handler.NewPersonalAccessTokenHandler(userRepo, patRepo).RegisterRoutes(internal)
	handler.NewAPILogHandler(apiLogRepo, userRepo).RegisterRoutes(internal)
	handler.NewDashboardHandler(userRepo, patRepo, apiLogRepo).RegisterRoutes(internal)
	handler.NewCompletionHandler(completionRepo, userRepo).RegisterRoutes(internal)

	aiStatus := ai.NewStatusTracker(config.Envs.GeminiApiKey != "", time.Duration(config.Envs.GeminiQuotaCooldown)*time.Second)
	aiProvider := ai.NewGeminiClient(config.Envs.GeminiApiKey, config.Envs.GeminiModel)
	aiService := ai.NewService(aiProvider, aiStatus)
	handler.NewAssistHandler(aiService, userRepo).RegisterRoutes(internal)
	handler.NewAccountHandler(userRepo).RegisterRoutes(internal)
	handler.NewSecurityHandler(userRepo, refreshRepo).RegisterRoutes(internal)
	handler.NewDataExportHandler(userRepo, profileRepo, skillRepo, educationRepo, experienceRepo, projectRepo, certificationRepo).RegisterRoutes(internal)

	public.NewProfileHandler(profileRepo, patRepo, apiLogRepo).RegisterRoutes(publicRouter)
	public.NewSkillHandler(skillRepo, patRepo, apiLogRepo).RegisterRoutes(publicRouter)
	public.NewEducationHandler(educationRepo, patRepo, apiLogRepo).RegisterRoutes(publicRouter)
	public.NewProjectHandler(projectRepo, patRepo, apiLogRepo).RegisterRoutes(publicRouter)
	public.NewExperienceHandler(experienceRepo, patRepo, apiLogRepo).RegisterRoutes(publicRouter)
	public.NewCertificationHandler(certificationRepo, patRepo, apiLogRepo).RegisterRoutes(publicRouter)

	corsRouter := CORS(router)

	log.Println("Listening on", s.addr)
	return http.ListenAndServe(s.addr, corsRouter)
}
