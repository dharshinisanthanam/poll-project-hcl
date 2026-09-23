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

	// Connect to MongoDB (non-fatal; will retry in background if initial ping fails)
	mongoRepo, err := repository.NewMongoRepo(cfg.MongoURI, cfg.MongoDB)
	if err != nil {
		log.Printf("================================================================================")
		log.Printf("WARNING: Initial connection to MongoDB failed: %v", err)
		log.Printf(">> The server will start and keep retrying connection in the background.")
		log.Printf(">> For Render Deployment: Please configure MONGO_URI in your dashboard settings")
		log.Printf("   with your MongoDB Atlas URI: mongodb+srv://<user>:<password>@cluster0.mongodb.net/livepolling?retryWrites=true&w=majority")
		log.Printf(">> For Local Development:")
		log.Printf("   Ensure MongoDB daemon is running locally: .\\tools\\mongo\\bin\\mongod.exe or run .\\run-local.ps1")
		log.Printf("================================================================================")
	} else {
		log.Printf("Connected successfully to MongoDB at %s", cfg.MongoURI)
	}
	defer mongoRepo.Close()

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
