package controllers

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"time"

	"live-polling/backend/models"
	"live-polling/backend/repository"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type PollController struct {
	mongoRepo *repository.MongoRepo
	redisRepo *repository.RedisRepo
}

func NewPollController(mongoRepo *repository.MongoRepo, redisRepo *repository.RedisRepo) *PollController {
	return &PollController{
		mongoRepo: mongoRepo,
		redisRepo: redisRepo,
	}
}

// Generate a random 6-character uppercase alphanumeric share code
func generateShareCode() string {
	const charset = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789" // avoid confusing characters like 0/O, 1/I
	b := make([]byte, 6)
	for i := range b {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return fmt.Sprintf("P%d", time.Now().UnixNano()%100000)
		}
		b[i] = charset[n.Int64()]
	}
	return string(b)
}

func (pc *PollController) CreatePoll(c *gin.Context) {
	userIDVal, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required to create a poll"})
		return
	}
	creatorID := userIDVal.(primitive.ObjectID)

	var req models.CreatePollRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	req.Question = strings.TrimSpace(req.Question)
	if len(req.Question) < 5 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Question must be at least 5 characters long"})
		return
	}

	if len(req.Options) < 2 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Poll must contain at least 2 options"})
		return
	}

	if len(req.Options) > 10 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Poll cannot contain more than 10 options"})
		return
	}

	// Validate options and check for duplicates
	seen := make(map[string]bool)
	var options []models.Option

	for i, optText := range req.Options {
		trimmed := strings.TrimSpace(optText)
		if trimmed == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Option %d cannot be empty", i+1)})
			return
		}
		lower := strings.ToLower(trimmed)
		if seen[lower] {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Duplicate option found: '%s'", trimmed)})
			return
		}
		seen[lower] = true

		options = append(options, models.Option{
			ID:    fmt.Sprintf("opt_%d", i+1),
			Text:  trimmed,
			Votes: 0,
		})
	}

	// Calculate expiration
	var expiresAt *time.Time
	now := time.Now()
	switch strings.ToLower(strings.TrimSpace(req.ExpiresIn)) {
	case "1h":
		t := now.Add(1 * time.Hour)
		expiresAt = &t
	case "1d":
		t := now.Add(24 * time.Hour)
		expiresAt = &t
	case "7d":
		t := now.Add(7 * 24 * time.Hour)
		expiresAt = &t
	default:
		// "never" or unspecified
		expiresAt = nil
	}

	shareCode := generateShareCode()

	poll := &models.Poll{
		Question:  req.Question,
		Options:   options,
		CreatorID: creatorID,
		ShareCode: shareCode,
		Status:    "active",
		ExpiresAt: expiresAt,
	}

	if err := pc.mongoRepo.CreatePoll(c.Request.Context(), poll); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create poll: " + err.Error()})
		return
	}

	// Initialize vote counts in Redis
	_ = pc.redisRepo.InitPollVotes(c.Request.Context(), poll.ID.Hex(), poll.Options)

	c.JSON(http.StatusCreated, gin.H{
		"message": "Poll created successfully",
		"poll":    poll,
	})
}

func (pc *PollController) GetUserPolls(c *gin.Context) {
	userIDVal, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}
	creatorID := userIDVal.(primitive.ObjectID)

	polls, err := pc.mongoRepo.GetPollsByCreatorID(c.Request.Context(), creatorID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user polls"})
		return
	}

	// Hydrate live counts from Redis
	for i := range polls {
		redisCounts, totalVotes, err := pc.redisRepo.GetPollVotes(c.Request.Context(), polls[i].ID.Hex())
		if err == nil && len(redisCounts) > 0 {
			var total int64 = 0
			for j := range polls[i].Options {
				if v, ok := redisCounts[polls[i].Options[j].ID]; ok {
					polls[i].Options[j].Votes = v
					total += v
				}
			}
			_ = totalVotes
		}
	}

	c.JSON(http.StatusOK, gin.H{"polls": polls})
}

func (pc *PollController) GetPollByShareCode(c *gin.Context) {
	shareCode := strings.ToUpper(strings.TrimSpace(c.Param("shareCode")))
	if shareCode == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Share code is required"})
		return
	}

	poll, err := pc.mongoRepo.GetPollByShareCode(c.Request.Context(), shareCode)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "Poll not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Check if expired
	now := time.Now()
	if poll.ExpiresAt != nil && now.After(*poll.ExpiresAt) && poll.Status == "active" {
		poll.Status = "closed"
		_ = pc.mongoRepo.UpdatePollStatus(c.Request.Context(), poll.ID, poll.CreatorID, "closed")
	}

	// Check if requester is creator
	isOwner := false
	if userIDVal, exists := c.Get("userID"); exists {
		if creatorID, ok := userIDVal.(primitive.ObjectID); ok && creatorID == poll.CreatorID {
			isOwner = true
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"poll":     poll,
		"is_owner": isOwner,
	})
}

func (pc *PollController) GetPollByID(c *gin.Context) {
	idStr := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid poll ID"})
		return
	}

	poll, err := pc.mongoRepo.GetPollByID(c.Request.Context(), objID)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "Poll not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Check if requester is creator
	isOwner := false
	if userIDVal, exists := c.Get("userID"); exists {
		if creatorID, ok := userIDVal.(primitive.ObjectID); ok && creatorID == poll.CreatorID {
			isOwner = true
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"poll":     poll,
		"is_owner": isOwner,
	})
}

func (pc *PollController) UpdatePollStatus(c *gin.Context) {
	userIDVal, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}
	creatorID := userIDVal.(primitive.ObjectID)

	idStr := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid poll ID"})
		return
	}

	var req models.UpdateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid status: must be 'active' or 'closed'"})
		return
	}

	if err := pc.mongoRepo.UpdatePollStatus(c.Request.Context(), objID, creatorID, req.Status); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	// Broadcast status update event via Redis
	poll, _ := pc.mongoRepo.GetPollByID(c.Request.Context(), objID)
	if poll != nil {
		redisCounts, totalVotes, _ := pc.redisRepo.GetPollVotes(c.Request.Context(), poll.ID.Hex())
		var optionResults []models.OptionResult
		for _, opt := range poll.Options {
			v := redisCounts[opt.ID]
			pct := 0.0
			if totalVotes > 0 {
				pct = (float64(v) / float64(totalVotes)) * 100.0
			}
			optionResults = append(optionResults, models.OptionResult{
				ID:         opt.ID,
				Text:       opt.Text,
				Votes:      v,
				Percentage: pct,
			})
		}
		_ = pc.redisRepo.PublishVoteUpdate(c.Request.Context(), poll.ID.Hex(), models.VoteEventMessage{
			Event:  "POLL_STATUS_CHANGED",
			PollID: poll.ID.Hex(),
			Payload: models.PollResultResponse{
				PollID:     poll.ID.Hex(),
				Question:   poll.Question,
				ShareCode:  poll.ShareCode,
				Status:     req.Status,
				TotalVotes: totalVotes,
				Options:    optionResults,
				ExpiresAt:  poll.ExpiresAt,
				CreatedAt:  poll.CreatedAt,
			},
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Poll status updated",
		"status":  req.Status,
	})
}

func (pc *PollController) DeletePoll(c *gin.Context) {
	userIDVal, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}
	creatorID := userIDVal.(primitive.ObjectID)

	idStr := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid poll ID"})
		return
	}

	if err := pc.mongoRepo.DeletePoll(c.Request.Context(), objID, creatorID); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Poll deleted successfully"})
}
