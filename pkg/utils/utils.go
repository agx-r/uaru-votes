package utils

import (
	"errors"
	"log/slog"
	"strconv"
	"strings"

	tb "gopkg.in/telebot.v4"
)

var ErrInvalidLogLevel = errors.New("provided level string is not one of debug/info/warn/error neither a valid int")

func ParseLogLevel(level string) (slog.Level, error) {
	level = strings.ToLower(strings.TrimSpace(level))
	switch level {
	case "debug":
		return slog.LevelDebug, nil
	case "info":
		return slog.LevelInfo, nil
	case "warn":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		levelInt, err := strconv.Atoi(level)
		if err != nil {
			var zeroLevel slog.Level

			return zeroLevel, ErrInvalidLogLevel
		}

		return slog.Level(levelInt), nil
	}
}

func ErrorAttr(err error) slog.Attr {
	return slog.String("err", err.Error())
}

type ErrorHandler struct {
	logger *slog.Logger
}

func NewErrorHandler(logger *slog.Logger) ErrorHandler {
	return ErrorHandler{logger: logger}
}

func (h *ErrorHandler) HandleError(err error, _ tb.Context) {
	if err != nil {
		h.logger.Error("error from bot", ErrorAttr(err))
	}
}

func IsAdmin(userID int64, admins []tb.ChatMember) bool {
	for _, admin := range admins {
		if admin.User.ID == userID {
			return true
		}
	}

	return false
}

func BotCanMute(botID int64, admins []tb.ChatMember) bool {
	for _, admin := range admins {
		if admin.User.ID == botID && admin.CanRestrictMembers {
			return true
		}
	}

	return false
}
