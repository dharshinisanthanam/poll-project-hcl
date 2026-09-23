package repository

import (
	"context"
	"errors"
	"log"
	"regexp"
	"strings"
	"sync"
	"time"

	"live-polling/backend/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoRepo struct {
	mu          sync.RWMutex
	client      *mongo.Client
	db          *mongo.Database
	users       *mongo.Collection
	polls       *mongo.Collection
	votes       *mongo.Collection
	isConnected bool
	lastError   string
	uri         string
	dbName      string
}

func NewMongoRepo(uri, dbName string) (*MongoRepo, error) {
	repo := &MongoRepo{
		uri:    uri,
		dbName: dbName,
	}

	err := repo.connect(8 * time.Second)
	if err != nil {
		repo.mu.Lock()
		repo.isConnected = false
		repo.lastError = err.Error()
		repo.mu.Unlock()

		log.Printf("================================================================================")
		log.Printf("[MongoDB] Initial connection attempt failed: %v", err)
		log.Printf("[MongoDB] Starting background reconnection loop (retrying every 5 seconds)...")
		log.Printf("================================================================================")

		go repo.backgroundReconnect()
		return repo, err
	}

	return repo, nil
}

func (r *MongoRepo) connect(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	clientOpts := options.Client().ApplyURI(r.uri)
	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		return err
	}

	if err := client.Ping(ctx, nil); err != nil {
		_ = client.Disconnect(ctx)
		return err
	}

	db := client.Database(r.dbName)
	r.mu.Lock()
	r.client = client
	r.db = db
	r.users = db.Collection("users")
	r.polls = db.Collection("polls")
	r.votes = db.Collection("votes")
	r.isConnected = true
	r.lastError = ""
	r.mu.Unlock()

	r.ensureIndexes()
	return nil
}

func (r *MongoRepo) backgroundReconnect() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		if r.IsConnected() {
			return
		}
		log.Printf("[MongoDB] Retrying connection to MongoDB Atlas...")
		if err := r.connect(8 * time.Second); err == nil {
			log.Printf("[MongoDB] >>> Successfully connected to MongoDB Atlas database '%s'! <<<", r.dbName)
			return
		} else {
			r.mu.Lock()
			r.lastError = err.Error()
			r.mu.Unlock()
			log.Printf("[MongoDB] Reconnection attempt failed: %v", err)
		}
	}
}

func (r *MongoRepo) IsConnected() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.isConnected && r.client != nil
}

func (r *MongoRepo) GetLastError() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.lastError
}

func (r *MongoRepo) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.isConnected = false
	if r.client != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return r.client.Disconnect(ctx)
	}
	return nil
}

func (r *MongoRepo) ensureIndexes() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	r.mu.RLock()
	users := r.users
	polls := r.polls
	votes := r.votes
	r.mu.RUnlock()

	if users == nil || polls == nil || votes == nil {
		return
	}

	// Unique email index for users
	_, _ = users.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.M{"email": 1},
		Options: options.Index().SetUnique(true),
	})

	// Unique share_code index for polls
	_, _ = polls.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.M{"share_code": 1},
		Options: options.Index().SetUnique(true),
	})

	// Compound unique index for votes: 1 vote per voter_ident per poll
	_, _ = votes.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "poll_id", Value: 1},
			{Key: "voter_ident", Value: 1},
		},
		Options: options.Index().SetUnique(true),
	})
}

// User methods
func (r *MongoRepo) CreateUser(ctx context.Context, user *models.User) error {
	if !r.IsConnected() {
		return errors.New("database connection is initializing")
	}
	user.ID = primitive.NewObjectID()
	user.CreatedAt = time.Now()
	_, err := r.users.InsertOne(ctx, user)
	return err
}

func (r *MongoRepo) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	if !r.IsConnected() {
		return nil, errors.New("database connection is initializing")
	}
	var user models.User
	err := r.users.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *MongoRepo) GetUserByEmailOrName(ctx context.Context, identifier string) (*models.User, error) {
	if !r.IsConnected() {
		return nil, errors.New("database connection is initializing")
	}
	var user models.User
	clean := strings.TrimSpace(identifier)
	filter := bson.M{
		"$or": []bson.M{
			{"email": strings.ToLower(clean)},
			{"name": primitive.Regex{Pattern: "^" + regexp.QuoteMeta(clean) + "$", Options: "i"}},
		},
	}
	err := r.users.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *MongoRepo) GetUserByID(ctx context.Context, id primitive.ObjectID) (*models.User, error) {
	if !r.IsConnected() {
		return nil, errors.New("database connection is initializing")
	}
	var user models.User
	err := r.users.FindOne(ctx, bson.M{"_id": id}).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Poll methods
func (r *MongoRepo) CreatePoll(ctx context.Context, poll *models.Poll) error {
	if !r.IsConnected() {
		return errors.New("database connection is initializing")
	}
	poll.ID = primitive.NewObjectID()
	poll.CreatedAt = time.Now()
	_, err := r.polls.InsertOne(ctx, poll)
	return err
}

func (r *MongoRepo) GetPollByID(ctx context.Context, id primitive.ObjectID) (*models.Poll, error) {
	if !r.IsConnected() {
		return nil, errors.New("database connection is initializing")
	}
	var poll models.Poll
	err := r.polls.FindOne(ctx, bson.M{"_id": id}).Decode(&poll)
	if err != nil {
		return nil, err
	}
	return &poll, nil
}

func (r *MongoRepo) GetPollByShareCode(ctx context.Context, shareCode string) (*models.Poll, error) {
	if !r.IsConnected() {
		return nil, errors.New("database connection is initializing")
	}
	var poll models.Poll
	err := r.polls.FindOne(ctx, bson.M{"share_code": shareCode}).Decode(&poll)
	if err != nil {
		return nil, err
	}
	return &poll, nil
}

func (r *MongoRepo) GetPollsByCreatorID(ctx context.Context, creatorID primitive.ObjectID) ([]models.Poll, error) {
	if !r.IsConnected() {
		return nil, errors.New("database connection is initializing")
	}
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	cursor, err := r.polls.Find(ctx, bson.M{"creator_id": creatorID}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var polls []models.Poll
	if err := cursor.All(ctx, &polls); err != nil {
		return nil, err
	}
	if polls == nil {
		polls = []models.Poll{}
	}
	return polls, nil
}

func (r *MongoRepo) UpdatePollStatus(ctx context.Context, id primitive.ObjectID, creatorID primitive.ObjectID, status string) error {
	if !r.IsConnected() {
		return errors.New("database connection is initializing")
	}
	res, err := r.polls.UpdateOne(ctx,
		bson.M{"_id": id, "creator_id": creatorID},
		bson.M{"$set": bson.M{"status": status}},
	)
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return errors.New("poll not found or unauthorized")
	}
	return nil
}

func (r *MongoRepo) DeletePoll(ctx context.Context, id primitive.ObjectID, creatorID primitive.ObjectID) error {
	if !r.IsConnected() {
		return errors.New("database connection is initializing")
	}
	res, err := r.polls.DeleteOne(ctx, bson.M{"_id": id, "creator_id": creatorID})
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return errors.New("poll not found or unauthorized")
	}
	// Also delete votes for this poll
	_, _ = r.votes.DeleteMany(ctx, bson.M{"poll_id": id})
	return nil
}

// Vote methods
func (r *MongoRepo) HasVoted(ctx context.Context, pollID primitive.ObjectID, voterIdent string) (bool, error) {
	if !r.IsConnected() {
		return false, errors.New("database connection is initializing")
	}
	count, err := r.votes.CountDocuments(ctx, bson.M{"poll_id": pollID, "voter_ident": voterIdent})
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *MongoRepo) RecordVote(ctx context.Context, pollID primitive.ObjectID, optionID string, voterIdent string) error {
	if !r.IsConnected() {
		return errors.New("database connection is initializing")
	}
	vote := models.VoteRecord{
		ID:         primitive.NewObjectID(),
		PollID:     pollID,
		OptionID:   optionID,
		VoterIdent: voterIdent,
		CreatedAt:  time.Now(),
	}
	_, err := r.votes.InsertOne(ctx, vote)
	return err
}

// Sync option vote counts in Mongo when needed
func (r *MongoRepo) UpdateOptionVoteCount(ctx context.Context, pollID primitive.ObjectID, optionID string) error {
	if !r.IsConnected() {
		return errors.New("database connection is initializing")
	}
	_, err := r.polls.UpdateOne(
		ctx,
		bson.M{"_id": pollID, "options.id": optionID},
		bson.M{"$inc": bson.M{"options.$.votes": 1}},
	)
	return err
}
