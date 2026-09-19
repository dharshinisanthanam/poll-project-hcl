package middleware

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func CORSMiddleware(frontendURL string) gin.HandlerFunc {
	config := cors.Config{
		AllowOrigins: []string{
			"http://localhost:5173",
			"http://localhost:3000",
			"http://127.0.0.1:5173",
			"http://127.0.0.1:3000",
		},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization", "Accept", "X-Requested-With"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}

	if frontendURL != "" && frontendURL != "http://localhost:5173" {
		config.AllowOrigins = append(config.AllowOrigins, frontendURL)
	}

	// Also allow all origins if deploying across dynamic domains (e.g. preview links)
	config.AllowOriginFunc = func(origin string) bool {
		return true // Allow all for seamless demo/testing
	}

	return cors.New(config)
}
