package config

import (
	"log"
	"os"
	"strconv"
)

type Config struct {
	PublicHost string
	Port       string

	DBUser                 string
	DBPassword             string
	DBHost                 string
	DBPort                 string
	DBName                 string
	JWTExpirationInSeconds int64
	JWTSecret              string
	R2AccountId            string
	R2AccessKeyId          string
	R2SecretAccessKey      string
	R2Bucket               string
	R2Endpoint             string
	R2PublicUrl            string
	BackendUrl             string
	FrontendUrl            string
	GoogleOauthClientId    string
	GoogleOauthSecret      string
	SmtpHost               string
	SmtpPort               int
	SmtpUsername           string
	SmtpPassword           string
	SmtpFrom               string
	SmtpEncryption         string
	GeminiApiKey           string
	GeminiModel            string
	GeminiQuotaCooldown    int
}
var Envs Config

func initConfig() Config {
	return Config{
		PublicHost: getEnv("PUBLIC_HOST", "http://localhost"),
		Port:       getEnv("PORT", "8080"),
		DBPort:     getEnv("DB_PORT", "3306"),
		DBUser:     getEnv("DB_USER", "root"),
		DBPassword: getEnv("DB_PASSWORD", ""),
		DBHost:     getEnv("DB_HOST", "127.0.0.1"),
		DBName:     getEnv("DB_NAME", "megome"),
		// JWTExpirationInSeconds: getEnvAsInt("JWT_EXP", 60*5),
		JWTExpirationInSeconds: getEnvAsInt64("JWT_EXP", 1800),
		JWTSecret:              getEnv("JWT_SECRET", ""),
		R2AccountId:            getEnv("R2_ACCOUNT_ID", "4ee86bb26d20c0c74970845960bec979"),
		R2AccessKeyId:          getEnv("R2_ACCESS_KEY_ID", "783e12a9c12ecd2c966fbbac42225c5d"),
		R2SecretAccessKey:      getEnv("R2_SECRET_ACCESS_KEY", "3140e4fdea0f3ad4099205c41caf4270478eceb7cfcb5a6183f3897b90c777d4"),
		R2Bucket:               getEnv("R2_BUCKET", "megome"),
		R2Endpoint:             getEnv("R2_ENDPOINT", "4ee86bb26d20c0c74970845960bec979.r2.cloudflarestorage.com"),
		R2PublicUrl:            getEnv("R2_PUBLIC_URL", "https://pub-8f00a57b78e742a3ac1da0446971e45d.r2.dev"),
		BackendUrl:             getEnv("BACKEND_URL", "http://localhost:8080"),
		FrontendUrl:            getEnv("FRONTEND_URL", "http://localhost:3001"),
		GoogleOauthClientId:    getEnv("GOOGLE_OAUTH_CLIENT_ID", ""),
		GoogleOauthSecret:      getEnv("GOOGLE_OAUTH_SECRET", ""),
		SmtpHost:               getEnv("SMTP_HOST", ""),
		SmtpPort:               getEnvAsInt("SMTP_PORT", 587),
		SmtpUsername:           getEnv("SMTP_USERNAME", ""),
		SmtpPassword:           getEnv("SMTP_PASSWORD", ""),
		SmtpFrom:               getEnv("SMTP_FROM", ""),
		SmtpEncryption:         getEnv("SMTP_ENCRYPTION", ""),
		GeminiApiKey:           getEnv("GEMINI_API_KEY", ""),
		GeminiModel:            getEnv("GEMINI_MODEL", "gemini-2.0-flash"),
		GeminiQuotaCooldown:    getEnvAsInt("GEMINI_QUOTA_COOLDOWN", 1800),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}

	return fallback
}

func getEnvAsInt64(key string, fallback int64) int64 {
	if value, ok := os.LookupEnv(key); ok {
		i, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return fallback
		}
		return i
	}
	return fallback
}

func getEnvAsInt(key string, fallback int) int {
	if value, ok := os.LookupEnv(key); ok {
		i, err := strconv.Atoi(value)
		if err != nil {
			return fallback
		}
		return i
	}
	return fallback
}

func validateConfig(cfg Config) {
	required := map[string]string{
		"JWT_SECRET":  cfg.JWTSecret,
		"DB_HOST":     cfg.DBHost,
		"DB_USER":     cfg.DBUser,
		"DB_PASSWORD": cfg.DBPassword,
		"DB_NAME":     cfg.DBName,
	}

	for key, value := range required {
		if value == "" {
			log.Fatalf("missing required env: %s", key)
		}
	}
}

func Load() {
	Envs = initConfig()
	validateConfig(Envs)
}
