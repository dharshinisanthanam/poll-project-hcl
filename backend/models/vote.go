package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type VoteRecord struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	PollID     primitive.ObjectID `bson:"poll_id" json:"poll_id"`
	OptionID   string             `bson:"option_id" json:"option_id"`
	VoterIdent string             `bson:"voter_ident" json:"voter_ident"` // sha256 of IP + VoterToken
	CreatedAt  time.Time          `bson:"created_at" json:"created_at"`
}

type VoteEventMessage struct {
	Event   string             `json:"event"` // "VOTE_UPDATED" or "POLL_STATUS_CHANGED"
	PollID  string             `json:"poll_id"`
	Payload PollResultResponse `json:"payload"`
}
