package config

import (
	"os"
	appErr "user-service/internal/errors"

	"github.com/joho/godotenv"
)

type Config struct {
	Port           string
	MongoURI       string
	DBName         string
	JWTSecret      string
	AuthServiceURL string
}

func LoadConfig() (*Config, error) {

	err := godotenv.Load()
	if err != nil {
		return nil, appErr.ErrInternalServer
	}

	cfg := &Config{
		Port:           getEnv("PORT", "8081"),
		MongoURI:       getEnv("MONGO_URI", "mongodb://localhost:27017"),
		DBName:         getEnv("DB_NAME", "userdb"),
		JWTSecret:      getEnv("JWT_SECRET", ""),
		AuthServiceURL: getEnv("AUTH_SERVICE_URL", ""),
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
