package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"live-polling/backend/config"
	"live-polling/backend/models"
	"live-polling/backend/repository"
	"live-polling/backend/routes"
	"live-polling/backend/websocket"

	"github.com/gin-gonic/gin"
)

func setupTestApp(t *testing.T) (*gin.Engine, *repository.MongoRepo, *repository.RedisRepo) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{
		Port:        "8080",
		MongoURI:    "mongodb://127.0.0.1:27017",
		MongoDB:     "livepolling_test",
		RedisURL:    "redis://127.0.0.1:6379",
		JWTSecret:   "test-secret-key-12345",
		FrontendURL: "http://localhost:5173",
	}

	mongoRepo, err := repository.NewMongoRepo(cfg.MongoURI, cfg.MongoDB)
	if err != nil {
		t.Fatalf("Failed to connect to MongoDB: %v", err)
	}

	redisRepo, err := repository.NewRedisRepo(cfg.RedisURL)
	if err != nil {
		t.Fatalf("Failed to connect to Redis: %v", err)
	}

	hub := websocket.NewHub(redisRepo)
	go hub.Run()

	router := routes.SetupRouter(cfg, mongoRepo, redisRepo, hub)
	return router, mongoRepo, redisRepo
}

func TestCompleteFlow(t *testing.T) {
	router, mongoRepo, redisRepo := setupTestApp(t)
	defer mongoRepo.Close()
	defer redisRepo.Close()

	// 1. Test User Signup
	signupPayload := models.SignupRequest{
		Name:     "Test Intern",
		Email:    "intern" + time.Now().Format("20060102150405") + "@guvi.in",
		Password: "password123",
	}
	body, _ := json.Marshal(signupPayload)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/signup", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("Expected status 201 Created, got %d: %s", w.Code, w.Body.String())
	}

	var signupRes struct {
		Message string              `json:"message"`
		User    models.UserResponse `json:"user"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &signupRes)
	token := signupRes.User.Token
	if token == "" {
		t.Fatalf("Expected token to be returned upon signup")
	}

	// 2. Test User Login
	loginPayload := models.LoginRequest{
		Email:    signupPayload.Email,
		Password: signupPayload.Password,
	}
	body, _ = json.Marshal(loginPayload)
	req = httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200 OK on login, got %d", w.Code)
	}

	// 3. Test Poll Creation Validation: Reject duplicate options
	invalidPoll := models.CreatePollRequest{
		Question: "What is your favorite language?",
		Options:  []string{"Go", "Go", "Python"},
	}
	body, _ = json.Marshal(invalidPoll)
	req = httptest.NewRequest(http.MethodPost, "/api/polls", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("Expected status 400 Bad Request on duplicate options, got %d", w.Code)
	}

	// 4. Test Poll Creation Validation: Reject < 2 options
	invalidPoll2 := models.CreatePollRequest{
		Question: "Only one option?",
		Options:  []string{"Solo"},
	}
	body, _ = json.Marshal(invalidPoll2)
	req = httptest.NewRequest(http.MethodPost, "/api/polls", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("Expected status 400 Bad Request on < 2 options, got %d", w.Code)
	}

	// 5. Test Valid Poll Creation
	validPoll := models.CreatePollRequest{
		Question:  "What is your favorite programming language?",
		Options:   []string{"Python", "JavaScript", "Go", "Java"},
		ExpiresIn: "never",
	}
	body, _ = json.Marshal(validPoll)
	req = httptest.NewRequest(http.MethodPost, "/api/polls", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("Expected status 201 Created on valid poll creation, got %d: %s", w.Code, w.Body.String())
	}

	var pollRes struct {
		Message string      `json:"message"`
		Poll    models.Poll `json:"poll"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &pollRes)
	pollID := pollRes.Poll.ID.Hex()
	shareCode := pollRes.Poll.ShareCode
	if pollID == "" || shareCode == "" {
		t.Fatalf("Expected valid poll ID and shareCode, got ID=%s, code=%s", pollID, shareCode)
	}

	// 6. Test Redis Pub/Sub subscription verification
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	pubsub := redisRepo.SubscribePollUpdates(ctx, pollID)
	defer pubsub.Close()
	pubsubCh := pubsub.Channel()

	// 7. Test Voting (Audience casting vote for "Go" -> opt_3)
	votePayload := models.VoteRequest{
		OptionID:   "opt_3",
		VoterToken: "voter-session-token-alpha-123",
	}
	body, _ = json.Marshal(votePayload)
	req = httptest.NewRequest(http.MethodPost, "/api/polls/"+shareCode+"/vote", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200 OK on vote, got %d: %s", w.Code, w.Body.String())
	}

	// 8. Verify Redis Pub/Sub received VOTE_UPDATED event
	select {
	case msg := <-pubsubCh:
		var event models.VoteEventMessage
		if err := json.Unmarshal([]byte(msg.Payload), &event); err != nil {
			t.Fatalf("Failed to parse Redis Pub/Sub event: %v", err)
		}
		if event.Event != "VOTE_UPDATED" {
			t.Fatalf("Expected event VOTE_UPDATED, got %s", event.Event)
		}
		if event.Payload.TotalVotes != 1 {
			t.Fatalf("Expected TotalVotes=1 in Pub/Sub payload, got %d", event.Payload.TotalVotes)
		}
		t.Logf("Successfully verified Redis Pub/Sub event: %s with TotalVotes=%d", event.Event, event.Payload.TotalVotes)
	case <-time.After(3 * time.Second):
		t.Fatalf("Timed out waiting for Redis Pub/Sub VOTE_UPDATED event")
	}

	// 9. Verify Redis Hash live vote counts directly
	redisCounts, totalVotes, err := redisRepo.GetPollVotes(context.Background(), pollID)
	if err != nil {
		t.Fatalf("Failed to read Redis vote counts: %v", err)
	}
	if totalVotes != 1 || redisCounts["opt_3"] != 1 {
		t.Fatalf("Redis counts mismatch: expected opt_3=1, total=1; got counts=%v, total=%d", redisCounts, totalVotes)
	}

	// 10. Test Duplicate Vote Prevention from same voter session
	req = httptest.NewRequest(http.MethodPost, "/api/polls/"+shareCode+"/vote", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Fatalf("Expected status 409 Conflict on duplicate vote, got %d: %s", w.Code, w.Body.String())
	}

	// 11. Test Live Results Endpoint
	req = httptest.NewRequest(http.MethodGet, "/api/polls/"+shareCode+"/results?voter_token=voter-session-token-alpha-123", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200 OK on results, got %d", w.Code)
	}

	var resultsRes struct {
		TotalVotes int64                 `json:"total_votes"`
		Options    []models.OptionResult `json:"options"`
		HasVoted   bool                  `json:"has_voted"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &resultsRes)

	if resultsRes.TotalVotes != 1 {
		t.Fatalf("Expected TotalVotes 1, got %d", resultsRes.TotalVotes)
	}
	if !resultsRes.HasVoted {
		t.Fatalf("Expected has_voted to be true for this voter")
	}

	t.Log("All 11 integration test assertions passed successfully!")
}
