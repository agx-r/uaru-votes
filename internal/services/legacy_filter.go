package services

import (
	"log/slog"
	"math/rand"
	"time"

	tb "gopkg.in/telebot.v4"
)

type LegacyMessageFilter struct {
	bot    tb.API
	logger *slog.Logger
}

func NewLegacyMessageFilter(bot tb.API, logger *slog.Logger) *LegacyMessageFilter {
	return &LegacyMessageFilter{
		bot:    bot,
		logger: logger,
	}
}

// удаляет сообщения КОЕ-КОГО
func (s *LegacyMessageFilter) HandleLegacyMessage(msg *tb.Message) error {
	// ID КОЕ-КОГО, для которого была создана эта логика
	const historicalUserID = 7952262321

	if msg.Sender.ID == historicalUserID {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		if r.Float64() < 0.6 {
			if err := s.bot.Delete(msg); err != nil {
				s.logger.Error("failed to delete message",
					slog.Int64("user_id", msg.Sender.ID),
					slog.Int64("message_id", int64(msg.ID)),
					slog.String("error", err.Error()))
			} else {
				s.logger.Info("message deleted",
					slog.Int64("user_id", msg.Sender.ID),
					slog.Int64("message_id", int64(msg.ID)))
			}
		}
	}

	return nil
}
