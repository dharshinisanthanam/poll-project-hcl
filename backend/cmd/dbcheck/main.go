package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"live-polling/backend/config"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/redis/go-redis/v9"
)

func main() {
	cfg := config.LoadConfig()

	fmt.Println("==================================================")
	fmt.Println("  MongoDB Database Connection & Verification Tool ")
	fmt.Println("==================================================")
	fmt.Printf("Attempting to connect to: %s\n", cfg.MongoURI)
	fmt.Printf("Target Database: %s\n\n", cfg.MongoDB)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.MongoURI))
	if err != nil {
		log.Fatalf("❌ FAILED to create MongoDB client: %v", err)
	}
	defer client.Disconnect(ctx)

	// Ping database
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatalf("❌ FAILED to ping MongoDB: %v\nCheck if MongoDB is running on %s", err, cfg.MongoURI)
	}

	fmt.Println("✅ CONNECTED SUCCESSFULLY to MongoDB!")

	// List databases
	dbList, err := client.ListDatabaseNames(ctx, bson.M{})
	if err == nil {
		fmt.Printf("Available Databases: %v\n\n", dbList)
	}

	db := client.Database(cfg.MongoDB)

	// List collections
	collections, err := db.ListCollectionNames(ctx, bson.M{})
	if err != nil {
		log.Fatalf("Failed to list collections: %v", err)
	}
	fmt.Printf("Collections in '%s': %v\n\n", cfg.MongoDB, collections)

	// Inspect Users
	usersColl := db.Collection("users")
	userCount, _ := usersColl.CountDocuments(ctx, bson.M{})
	fmt.Printf("--- Users Collection (%d documents) ---\n", userCount)
	cursor, _ := usersColl.Find(ctx, bson.M{}, options.Find().SetLimit(5))
	var users []bson.M
	_ = cursor.All(ctx, &users)
	for i, u := range users {
		// hide password hash
		delete(u, "password_hash")
		b, _ := json.MarshalIndent(u, "", "  ")
		fmt.Printf("[%d] %s\n", i+1, string(b))
	}

	// Inspect Polls
	pollsColl := db.Collection("polls")
	pollCount, _ := pollsColl.CountDocuments(ctx, bson.M{})
	fmt.Printf("\n--- Polls Collection (%d documents) ---\n", pollCount)
	cursor, _ = pollsColl.Find(ctx, bson.M{}, options.Find().SetLimit(5))
	var polls []bson.M
	_ = cursor.All(ctx, &polls)
	for i, p := range polls {
		b, _ := json.MarshalIndent(p, "", "  ")
		fmt.Printf("[%d] %s\n", i+1, string(b))
	}

	// Inspect Votes
	votesColl := db.Collection("votes")
	voteCount, _ := votesColl.CountDocuments(ctx, bson.M{})
	fmt.Printf("\n--- Votes Collection (%d records) ---\n", voteCount)

	// Verify Redis Connection & Live Hashes
	fmt.Println("\n==================================================")
	fmt.Println("  Redis Database Connection & Verification        ")
	fmt.Println("==================================================")
	fmt.Printf("Attempting to connect to Redis at: %s\n", cfg.RedisURL)

	opts, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		opts = &redis.Options{Addr: cfg.RedisURL}
	}
	rClient := redis.NewClient(opts)
	defer rClient.Close()

	rCtx, rCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer rCancel()

	if err := rClient.Ping(rCtx).Err(); err != nil {
		log.Fatalf("❌ FAILED to connect to Redis: %v\nCheck if Redis is running on %s", err, cfg.RedisURL)
	}
	fmt.Println("✅ CONNECTED SUCCESSFULLY to Redis!")

	keys, _ := rClient.Keys(rCtx, "poll:*:votes").Result()
	fmt.Printf("Active Redis Poll Hashes (%d active polls in Redis):\n", len(keys))
	for i, k := range keys {
		if i < 5 {
			counts, _ := rClient.HGetAll(rCtx, k).Result()
			fmt.Printf("  [%d] %s => %v\n", i+1, k, counts)
		}
	}

	fmt.Println("\n==================================================")
	fmt.Println("  BOTH MongoDB & Redis are 100% CONNECTED & ACTIVE! ")
	fmt.Println("==================================================")
}
