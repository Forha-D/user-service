package config

import (
	"os"
	"strconv"
	"time"
	appErr "user-service/internal/errors"

	"github.com/joho/godotenv"
)

type Config struct {
	Port           string
	MongoURI       string
	DBName         string
	JWTSecret      string
	AuthServiceURL string
	DBTimeout      time.Duration
}

func LoadConfig() (*Config, error) {

	err := godotenv.Load()
	if err != nil {
		return nil, appErr.ErrInternalServer
	}

	defaultTimeout := 5
	if timeoutStr := getEnv("DB_TIMEOUT", "5"); timeoutStr != "" {
		parsed, err := strconv.Atoi(timeoutStr)
		if err != nil || parsed <= 0 {
			return nil, appErr.ErrInvalidConfig
		}
		defaultTimeout = parsed
	}

	cfg := &Config{
		Port:           getEnv("PORT", "8081"),
		MongoURI:       getEnv("MONGO_URI", "mongodb://localhost:27017"),
		DBName:         getEnv("DB_NAME", "userdb"),
		JWTSecret:      getEnv("JWT_SECRET", ""),
		AuthServiceURL: getEnv("AUTH_SERVICE_URL", ""),
		DBTimeout:      time.Duration(defaultTimeout) * time.Second,
	}

	if cfg.JWTSecret == "" {
		return nil, appErr.ErrMissingEnv
	}

	if cfg.AuthServiceURL == "" {
		return nil, appErr.ErrMissingEnv
	}

	return cfg, nil
}

func getEnv(key string, fallback string) string {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	return val
}
