package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	PublicHost string
	Port       string

	DBUser                 string
	DBPassword             string
	DBAddress              string
	DBName                 string
	JWTExpirationInSeconds int64
	JWTSecret              string
	R2AccountId            string
	R2AccessKeyId          string
	R2SecretAccessKey      string
	R2Bucket               string
	R2Endpoint             string
	R2PublicUrl            string
}

var Envs Config

func initConfig() Config {
	return Config{
		PublicHost: getEnv("PUBLIC_HOST", "http://localhost"),
		Port:       getEnv("PORT", "3306"),
		DBUser:     getEnv("DB_USER", "root"),
		DBPassword: getEnv("DB_PASSWORD", ""),
		DBAddress:  fmt.Sprintf("%s:%s", getEnv("DB_HOST", "127.0.0.1"), getEnv("DB_PORT", "3306")),
		DBName:     getEnv("DB_NAME", "megome"),
		// JWTExpirationInSeconds: getEnvAsInt("JWT_EXP", 60*5),
		JWTExpirationInSeconds: getEnvAsInt("JWT_EXP", 60),
		JWTSecret:              getEnv("JWT_SECRET", "not-secret-anymore?"),
		R2AccountId:            getEnv("R2_ACCOUNT_ID", "4ee86bb26d20c0c74970845960bec979"),
		R2AccessKeyId:          getEnv("R2_ACCESS_KEY_ID", "783e12a9c12ecd2c966fbbac42225c5d"),
		R2SecretAccessKey:      getEnv("R2_SECRET_ACCESS_KEY", "3140e4fdea0f3ad4099205c41caf4270478eceb7cfcb5a6183f3897b90c777d4"),
		R2Bucket:               getEnv("R2_BUCKET", "megome"),
		R2Endpoint:             getEnv("R2_ENDPOINT", "4ee86bb26d20c0c74970845960bec979.r2.cloudflarestorage.com"),
		R2PublicUrl:            getEnv("R2_PUBLIC_URL", "https://pub-8f00a57b78e742a3ac1da0446971e45d.r2.dev"),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}

	return fallback
}

func getEnvAsInt(key string, fallback int64) int64 {
	if value, ok := os.LookupEnv(key); ok {
		i, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return fallback
		}
		return i
	}
	return fallback
}

func Load() {
	Envs = initConfig()
}
