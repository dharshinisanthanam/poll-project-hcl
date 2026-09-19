package repository

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"time"

	"live-polling/backend/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoRepo struct {
	client *mongo.Client
	db     *mongo.Database
	users  *mongo.Collection
	polls  *mongo.Collection
	votes  *mongo.Collection
}

func NewMongoRepo(uri, dbName string) (*MongoRepo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOpts := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		return nil, err
	}

	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}

	db := client.Database(dbName)
	repo := &MongoRepo{
		client: client,
		db:     db,
		users:  db.Collection("users"),
		polls:  db.Collection("polls"),
		votes:  db.Collection("votes"),
	}

	// Setup indexes
	repo.ensureIndexes()

	return repo, nil
}

func (r *MongoRepo) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return r.client.Disconnect(ctx)
}

func (r *MongoRepo) ensureIndexes() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Unique email index for users
	_, _ = r.users.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.M{"email": 1},
		Options: options.Index().SetUnique(true),
	})

	// Unique share_code index for polls
	_, _ = r.polls.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.M{"share_code": 1},
		Options: options.Index().SetUnique(true),
	})

	// Compound unique index for votes: 1 vote per voter_ident per poll
	_, _ = r.votes.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "poll_id", Value: 1},
			{Key: "voter_ident", Value: 1},
		},
		Options: options.Index().SetUnique(true),
	})
}

// User methods
func (r *MongoRepo) CreateUser(ctx context.Context, user *models.User) error {
	user.ID = primitive.NewObjectID()
	user.CreatedAt = time.Now()
	_, err := r.users.InsertOne(ctx, user)
	return err
}

func (r *MongoRepo) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := r.users.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *MongoRepo) GetUserByEmailOrName(ctx context.Context, identifier string) (*models.User, error) {
	var user models.User
	clean := strings.TrimSpace(identifier)
	// Query matching email (case-insensitive) OR name (case-insensitive)
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
	var user models.User
	err := r.users.FindOne(ctx, bson.M{"_id": id}).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Poll methods
func (r *MongoRepo) CreatePoll(ctx context.Context, poll *models.Poll) error {
	poll.ID = primitive.NewObjectID()
	poll.CreatedAt = time.Now()
	_, err := r.polls.InsertOne(ctx, poll)
	return err
}

func (r *MongoRepo) GetPollByID(ctx context.Context, id primitive.ObjectID) (*models.Poll, error) {
	var poll models.Poll
	err := r.polls.FindOne(ctx, bson.M{"_id": id}).Decode(&poll)
	if err != nil {
		return nil, err
	}
	return &poll, nil
}

func (r *MongoRepo) GetPollByShareCode(ctx context.Context, shareCode string) (*models.Poll, error) {
	var poll models.Poll
	err := r.polls.FindOne(ctx, bson.M{"share_code": shareCode}).Decode(&poll)
	if err != nil {
		return nil, err
	}
	return &poll, nil
}

func (r *MongoRepo) GetPollsByCreatorID(ctx context.Context, creatorID primitive.ObjectID) ([]models.Poll, error) {
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
	count, err := r.votes.CountDocuments(ctx, bson.M{"poll_id": pollID, "voter_ident": voterIdent})
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *MongoRepo) RecordVote(ctx context.Context, pollID primitive.ObjectID, optionID string, voterIdent string) error {
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
	_, err := r.polls.UpdateOne(
		ctx,
		bson.M{"_id": pollID, "options.id": optionID},
		bson.M{"$inc": bson.M{"options.$.votes": 1}},
	)
	return err
}
