package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Option struct {
	ID    string `bson:"id" json:"id"`
	Text  string `bson:"text" json:"text"`
	Votes int64  `bson:"votes" json:"votes"`
}

type Poll struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Question  string             `bson:"question" json:"question"`
	Options   []Option           `bson:"options" json:"options"`
	CreatorID primitive.ObjectID `bson:"creator_id" json:"creator_id"`
	ShareCode string             `bson:"share_code" json:"share_code"`
	Status    string             `bson:"status" json:"status"` // "active" or "closed"
	ExpiresAt *time.Time         `bson:"expires_at,omitempty" json:"expires_at,omitempty"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
}

type CreatePollRequest struct {
	Question  string   `json:"question" binding:"required,min=5,max=500"`
	Options   []string `json:"options" binding:"required,min=2,max=10"`
	ExpiresIn string   `json:"expires_in"` // "1h", "1d", "7d", or "never"
}

type VoteRequest struct {
	OptionID   string `json:"option_id" binding:"required"`
	VoterToken string `json:"voter_token"` // client generated unique token stored in localStorage
}

type UpdateStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=active closed"`
}

type OptionResult struct {
	ID         string  `json:"id"`
	Text       string  `json:"text"`
	Votes      int64   `json:"votes"`
	Percentage float64 `json:"percentage"`
}

type PollResultResponse struct {
	PollID     string         `json:"poll_id"`
	Question   string         `json:"question"`
	ShareCode  string         `json:"share_code"`
	Status     string         `json:"status"`
	TotalVotes int64          `json:"total_votes"`
	Options    []OptionResult `json:"options"`
	ExpiresAt  *time.Time     `json:"expires_at,omitempty"`
	CreatedAt  time.Time      `json:"created_at"`
	IsOwner    bool           `json:"is_owner"`
}
