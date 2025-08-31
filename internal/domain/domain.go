package domain

import (
	"log/slog"
	"time"

	tb "gopkg.in/telebot.v4"
)

type Context interface {
	tb.Context
	BotUser() *tb.User

	Log() *slog.Logger
	WithLogger(*slog.Logger) Context
	PollStorage() PollStorage
	StartPollMonitoring(*ActivePoll)
	BotAPI() tb.API
}

type PollType string

const (
	PollTypeBan     PollType = "ban"
	PollTypeUnban   PollType = "unban"
	PollTypeGifs    PollType = "gifs"
	PollTypeMedia   PollType = "media"
)

// active poll that needs to be monitored
type ActivePoll struct {
	ID           string    `json:"id"`
	Type         PollType  `json:"type"`
	ChatID       int64     `json:"chat_id"`
	MessageID    int       `json:"message_id"`
	UserID       int64     `json:"user_id"`
	CreatedAt    time.Time `json:"created_at"`
	ExpiresAt    time.Time `json:"expires_at"`
	MemberData   []byte    `json:"member_data"`
}

type PollStorage interface {
	SavePoll(poll *ActivePoll) error
	GetPolls() ([]*ActivePoll, error)
	DeletePoll(id string) error
	GetPollsByType(pollType PollType) ([]*ActivePoll, error)
}
