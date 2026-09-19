package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port        string
	MongoURI    string
	MongoDB     string
	RedisURL    string
	JWTSecret   string
	FrontendURL string
}

func LoadConfig() *Config {
	// Attempt to load .env if it exists
	_ = godotenv.Load()

	port := getEnv("PORT", "8080")
	mongoURI := getEnv("MONGO_URI", "mongodb://127.0.0.1:27017")
	mongoDB := getEnv("MONGO_DB", "livepolling")
	redisURL := getEnv("REDIS_URL", "redis://127.0.0.1:6379")
	jwtSecret := getEnv("JWT_SECRET", "guvi-live-polling-developer-internship-jwt-secret-key-2026")
	frontendURL := getEnv("FRONTEND_URL", "http://localhost:5173")

	return &Config{
		Port:        port,
		MongoURI:    mongoURI,
		MongoDB:     mongoDB,
		RedisURL:    redisURL,
		JWTSecret:   jwtSecret,
		FrontendURL: frontendURL,
	}
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
