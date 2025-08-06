package domain

import (
	"log/slog"

	tb "gopkg.in/telebot.v4"
)

type Context interface {
	tb.Context
	BotUser() *tb.User

	Log() *slog.Logger
	WithLogger(*slog.Logger) Context
}
