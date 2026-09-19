package routes

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"live-polling/backend/config"
	"live-polling/backend/controllers"
	"live-polling/backend/middleware"
	"live-polling/backend/repository"
	"live-polling/backend/websocket"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func SetupRouter(
	cfg *config.Config,
	mongoRepo *repository.MongoRepo,
	redisRepo *repository.RedisRepo,
	hub *websocket.Hub,
) *gin.Engine {
	r := gin.Default()

	// Global Middlewares
	r.Use(middleware.CORSMiddleware(cfg.FrontendURL))

	// Locate frontend dist directory for unified full-stack serving
	distDir := ""
	candidates := []string{
		"/app/frontend/dist",
		"../frontend/dist",
		"./frontend/dist",
		"e:/pollhcl/frontend/dist",
	}
	for _, candidate := range candidates {
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			distDir = candidate
			break
		}
	}

	if distDir != "" {
		assetsDir := filepath.Join(distDir, "assets")
		if info, err := os.Stat(assetsDir); err == nil && info.IsDir() {
			r.Static("/assets", assetsDir)
		}
	}

	// Controllers
	authCtrl := controllers.NewAuthController(mongoRepo, cfg.JWTSecret)
	pollCtrl := controllers.NewPollController(mongoRepo, redisRepo)
	voteCtrl := controllers.NewVoteController(mongoRepo, redisRepo)

	// Root endpoint: Serves React SPA if built, or API status
	r.GET("/", func(c *gin.Context) {
		if distDir != "" {
			indexFile := filepath.Join(distDir, "index.html")
			if _, err := os.Stat(indexFile); err == nil {
				c.File(indexFile)
				return
			}
		}
		c.JSON(http.StatusOK, gin.H{
			"status":    "online",
			"service":   "Live Polling Backend API",
			"version":   "1.0.0",
			"timestamp": time.Now(),
			"services": gin.H{
				"mongodb": true,
				"redis":   true,
			},
			"frontend_url": cfg.FrontendURL,
		})
	})

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()

		redisPing := redisRepo.GetClient().Ping(ctx).Err() == nil

		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"timestamp": time.Now(),
			"services": gin.H{
				"mongodb": true,
				"redis":   redisPing,
			},
		})
	})

	api := r.Group("/api")
	{
		api.GET("", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"status":    "online",
				"message":   "Live Polling API endpoints are active",
				"timestamp": time.Now(),
			})
		})
		// Auth routes
		auth := api.Group("/auth")
		{
			auth.POST("/signup", authCtrl.Signup)
			auth.POST("/login", authCtrl.Login)
			auth.GET("/me", middleware.AuthMiddleware(cfg.JWTSecret), authCtrl.Me)
		}

		// Poll management routes
		polls := api.Group("/polls")
		{
			// Public / Audience endpoints
			polls.GET("/share/:shareCode", middleware.OptionalAuthMiddleware(cfg.JWTSecret), pollCtrl.GetPollByShareCode)
			polls.GET("/:id", middleware.OptionalAuthMiddleware(cfg.JWTSecret), pollCtrl.GetPollByID)
			polls.GET("/:id/results", middleware.OptionalAuthMiddleware(cfg.JWTSecret), voteCtrl.GetResults)
			polls.POST("/:id/vote", voteCtrl.CastVote)

			// Real-time WebSocket live updates
			polls.GET("/:id/live", func(c *gin.Context) {
				idParam := c.Param("id")

				// Resolve ID if given as ShareCode
				pollID := idParam
				if _, err := primitive.ObjectIDFromHex(idParam); err != nil {
					poll, err := mongoRepo.GetPollByShareCode(c.Request.Context(), idParam)
					if err == nil && poll != nil {
						pollID = poll.ID.Hex()
					}
				}

				hub.ServeWs(c.Writer, c.Request, pollID)
			})

			// Authenticated creator endpoints
			authorized := polls.Group("")
			authorized.Use(middleware.AuthMiddleware(cfg.JWTSecret))
			{
				authorized.POST("", pollCtrl.CreatePoll)
				authorized.GET("", pollCtrl.GetUserPolls)
				authorized.PATCH("/:id/status", pollCtrl.UpdatePollStatus)
				authorized.DELETE("/:id", pollCtrl.DeletePoll)
			}
		}
	}

	// WebSocket alias route at /ws/polls/:id
	r.GET("/ws/polls/:id", func(c *gin.Context) {
		idParam := c.Param("id")
		pollID := idParam
		if _, err := primitive.ObjectIDFromHex(idParam); err != nil {
			poll, err := mongoRepo.GetPollByShareCode(c.Request.Context(), idParam)
			if err == nil && poll != nil {
				pollID = poll.ID.Hex()
			}
		}
		hub.ServeWs(c.Writer, c.Request, pollID)
	})

	// Fallback route: static assets or React SPA navigation
	r.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path
		if strings.HasPrefix(path, "/api") || strings.HasPrefix(path, "/ws") {
			c.JSON(http.StatusNotFound, gin.H{"error": "API route not found"})
			return
		}
		if distDir != "" {
			// Check if file exists directly in dist (e.g. /favicon.svg, /icons.svg)
			filePath := filepath.Join(distDir, filepath.Clean(path))
			if info, err := os.Stat(filePath); err == nil && !info.IsDir() {
				c.File(filePath)
				return
			}
			// Fallback to index.html for client-side routing
			indexFile := filepath.Join(distDir, "index.html")
			if _, err := os.Stat(indexFile); err == nil {
				c.File(indexFile)
				return
			}
		}
		c.JSON(http.StatusNotFound, gin.H{"error": "Page not found"})
	})

	return r
}
