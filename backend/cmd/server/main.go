package main

import (
	"log"

	"live-polling/backend/config"
	"live-polling/backend/repository"
	"live-polling/backend/routes"
	"live-polling/backend/websocket"
)

func main() {
	cfg := config.LoadConfig()

	log.Printf("Starting Live Polling Service on port %s...", cfg.Port)

	// Connect to MongoDB
	mongoRepo, err := repository.NewMongoRepo(cfg.MongoURI, cfg.MongoDB)
	if err != nil {
		log.Fatalf("CRITICAL: Failed to connect to MongoDB at %s: %v", cfg.MongoURI, err)
	}
	defer mongoRepo.Close()
	log.Printf("Connected successfully to MongoDB at %s", cfg.MongoURI)

	// Connect to Redis
	redisRepo, err := repository.NewRedisRepo(cfg.RedisURL)
	if err != nil {
		log.Fatalf("CRITICAL: Failed to connect to Redis at %s: %v", cfg.RedisURL, err)
	}
	defer redisRepo.Close()
	log.Printf("Connected successfully to Redis at %s", cfg.RedisURL)

	// Initialize WebSocket Hub
	hub := websocket.NewHub(redisRepo)
	go hub.Run()
	log.Println("WebSocket Hub initialized with Redis Pub/Sub dispatcher")

	// Setup Gin Router & Routes
	router := routes.SetupRouter(cfg, mongoRepo, redisRepo, hub)

	log.Printf("Live Polling API listening on http://0.0.0.0:%s", cfg.Port)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Failed to run HTTP server: %v", err)
	}
}
