package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"live-polling/backend/models"

	"github.com/redis/go-redis/v9"
)

type RedisRepo struct {
	client *redis.Client
}

func NewRedisRepo(redisURL string) (*RedisRepo, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		// Fallback for raw host:port
		opts = &redis.Options{
			Addr: redisURL,
		}
	}

	client := redis.NewClient(opts)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	return &RedisRepo{client: client}, nil
}

func (r *RedisRepo) Close() error {
	return r.client.Close()
}

func (r *RedisRepo) GetClient() *redis.Client {
	return r.client
}

// Key format helpers
func (r *RedisRepo) pollVotesKey(pollID string) string {
	return fmt.Sprintf("poll:%s:votes", pollID)
}

func (r *RedisRepo) pollChannelName(pollID string) string {
	return fmt.Sprintf("poll:%s:updates", pollID)
}

// Initialize Redis hash for poll options with 0 votes
func (r *RedisRepo) InitPollVotes(ctx context.Context, pollID string, options []models.Option) error {
	key := r.pollVotesKey(pollID)
	pipe := r.client.Pipeline()
	for _, opt := range options {
		pipe.HSetNX(ctx, key, opt.ID, opt.Votes)
	}
	// Expire after 30 days of inactivity
	pipe.Expire(ctx, key, 30*24*time.Hour)
	_, err := pipe.Exec(ctx)
	return err
}

// Increment vote for option atomically via HINCRBY
func (r *RedisRepo) IncrementVote(ctx context.Context, pollID string, optionID string) (map[string]int64, int64, error) {
	key := r.pollVotesKey(pollID)

	// Atomic increment
	_, err := r.client.HIncrBy(ctx, key, optionID, 1).Result()
	if err != nil {
		return nil, 0, err
	}

	// Fetch all updated counts from Redis Hash
	return r.GetPollVotes(ctx, pollID)
}

// Retrieve vote counts directly from Redis Hash
func (r *RedisRepo) GetPollVotes(ctx context.Context, pollID string) (map[string]int64, int64, error) {
	key := r.pollVotesKey(pollID)
	vals, err := r.client.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, 0, err
	}

	counts := make(map[string]int64)
	var total int64 = 0

	for optID, strVal := range vals {
		num, _ := strconv.ParseInt(strVal, 10, 64)
		counts[optID] = num
		total += num
	}

	return counts, total, nil
}

// Publish real-time vote update event to Redis channel
func (r *RedisRepo) PublishVoteUpdate(ctx context.Context, pollID string, event models.VoteEventMessage) error {
	channel := r.pollChannelName(pollID)
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}
	return r.client.Publish(ctx, channel, data).Err()
}

// Subscribe to poll updates channel
func (r *RedisRepo) SubscribePollUpdates(ctx context.Context, pollID string) *redis.PubSub {
	channel := r.pollChannelName(pollID)
	return r.client.Subscribe(ctx, channel)
}
