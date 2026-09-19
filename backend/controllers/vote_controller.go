package controllers

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"

	"live-polling/backend/models"
	"live-polling/backend/repository"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type VoteController struct {
	mongoRepo *repository.MongoRepo
	redisRepo *repository.RedisRepo
}

func NewVoteController(mongoRepo *repository.MongoRepo, redisRepo *repository.RedisRepo) *VoteController {
	return &VoteController{
		mongoRepo: mongoRepo,
		redisRepo: redisRepo,
	}
}

// Generate unique identifier combining client IP and client voter token
func generateVoterIdent(c *gin.Context, voterToken string, pollID string) string {
	ip := c.ClientIP()
	raw := fmt.Sprintf("%s|%s|%s", ip, voterToken, pollID)
	hash := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(hash[:])
}

func (vc *VoteController) CastVote(c *gin.Context) {
	idParam := c.Param("id")

	// Resolve poll either by ObjectID or by ShareCode
	var poll *models.Poll
	var err error
	if objID, parseErr := primitive.ObjectIDFromHex(idParam); parseErr == nil {
		poll, err = vc.mongoRepo.GetPollByID(c.Request.Context(), objID)
	} else {
		poll, err = vc.mongoRepo.GetPollByShareCode(c.Request.Context(), idParam)
	}

	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "Poll not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Check if poll is active
	if poll.Status != "active" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "This poll is closed and no longer accepting votes"})
		return
	}

	// Check if poll is expired
	now := time.Now()
	if poll.ExpiresAt != nil && now.After(*poll.ExpiresAt) {
		poll.Status = "closed"
		_ = vc.mongoRepo.UpdatePollStatus(c.Request.Context(), poll.ID, poll.CreatorID, "closed")
		c.JSON(http.StatusBadRequest, gin.H{"error": "This poll has expired"})
		return
	}

	var req models.VoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid vote request: option_id is required"})
		return
	}

	// Validate option exists in poll
	optionValid := false
	for _, opt := range poll.Options {
		if opt.ID == req.OptionID {
			optionValid = true
			break
		}
	}
	if !optionValid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid option selected for this poll"})
		return
	}

	// Deduplication: prevent multiple votes from same voter/session
	voterIdent := generateVoterIdent(c, req.VoterToken, poll.ID.Hex())
	hasVoted, err := vc.mongoRepo.HasVoted(c.Request.Context(), poll.ID, voterIdent)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error checking voting status"})
		return
	}
	if hasVoted {
		c.JSON(http.StatusConflict, gin.H{"error": "You have already voted in this poll"})
		return
	}

	// 1. Persist vote in MongoDB
	if err := vc.mongoRepo.RecordVote(c.Request.Context(), poll.ID, req.OptionID, voterIdent); err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Vote could not be recorded (duplicate vote detected)"})
		return
	}
	_ = vc.mongoRepo.UpdateOptionVoteCount(c.Request.Context(), poll.ID, req.OptionID)

	// 2. Increment in Redis atomically via HINCRBY
	redisCounts, totalVotes, err := vc.redisRepo.IncrementVote(c.Request.Context(), poll.ID.Hex(), req.OptionID)
	if err != nil {
		// If Redis key wasn't initialized, initialize and retry
		_ = vc.redisRepo.InitPollVotes(c.Request.Context(), poll.ID.Hex(), poll.Options)
		redisCounts, totalVotes, _ = vc.redisRepo.IncrementVote(c.Request.Context(), poll.ID.Hex(), req.OptionID)
	}

	// 3. Build updated live result response
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

	resultPayload := models.PollResultResponse{
		PollID:     poll.ID.Hex(),
		Question:   poll.Question,
		ShareCode:  poll.ShareCode,
		Status:     poll.Status,
		TotalVotes: totalVotes,
		Options:    optionResults,
		ExpiresAt:  poll.ExpiresAt,
		CreatedAt:  poll.CreatedAt,
	}

	// 4. Publish update event to Redis Pub/Sub channel
	_ = vc.redisRepo.PublishVoteUpdate(c.Request.Context(), poll.ID.Hex(), models.VoteEventMessage{
		Event:   "VOTE_UPDATED",
		PollID:  poll.ID.Hex(),
		Payload: resultPayload,
	})

	c.JSON(http.StatusOK, gin.H{
		"message": "Vote recorded successfully",
		"results": resultPayload,
	})
}

func (vc *VoteController) GetResults(c *gin.Context) {
	idParam := c.Param("id")

	var poll *models.Poll
	var err error
	if objID, parseErr := primitive.ObjectIDFromHex(idParam); parseErr == nil {
		poll, err = vc.mongoRepo.GetPollByID(c.Request.Context(), objID)
	} else {
		poll, err = vc.mongoRepo.GetPollByShareCode(c.Request.Context(), idParam)
	}

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
		_ = vc.mongoRepo.UpdatePollStatus(c.Request.Context(), poll.ID, poll.CreatorID, "closed")
	}

	// Read fast live counts from Redis
	redisCounts, totalVotes, err := vc.redisRepo.GetPollVotes(c.Request.Context(), poll.ID.Hex())
	if err != nil || len(redisCounts) == 0 {
		// Fallback to MongoDB stored counts
		_ = vc.redisRepo.InitPollVotes(c.Request.Context(), poll.ID.Hex(), poll.Options)
		redisCounts, totalVotes, _ = vc.redisRepo.GetPollVotes(c.Request.Context(), poll.ID.Hex())
	}

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

	// Check if this viewer already voted
	voterToken := c.Query("voter_token")
	hasVoted := false
	if voterToken != "" {
		voterIdent := generateVoterIdent(c, voterToken, poll.ID.Hex())
		hasVoted, _ = vc.mongoRepo.HasVoted(c.Request.Context(), poll.ID, voterIdent)
	}

	// Check if user is owner
	isOwner := false
	if userIDVal, exists := c.Get("userID"); exists {
		if creatorID, ok := userIDVal.(primitive.ObjectID); ok && creatorID == poll.CreatorID {
			isOwner = true
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"poll_id":     poll.ID.Hex(),
		"question":    poll.Question,
		"share_code":  poll.ShareCode,
		"status":      poll.Status,
		"total_votes": totalVotes,
		"options":     optionResults,
		"expires_at":  poll.ExpiresAt,
		"created_at":  poll.CreatedAt,
		"has_voted":   hasVoted,
		"is_owner":    isOwner,
	})
}
