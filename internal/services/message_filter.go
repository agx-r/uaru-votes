package services

import (
	"log/slog"
	"math/rand"
	"os"
	"strconv"
	"time"

	tb "gopkg.in/telebot.v4"
)

type MessageFilterService struct {
	bot               tb.API
	logger            *slog.Logger
	targetUserID      int64
	deletionProbability float64
	legacyFilter      *LegacyMessageFilter
}

func NewMessageFilterService(bot tb.API, logger *slog.Logger) *MessageFilterService {
	service := &MessageFilterService{
		bot:                 bot,
		logger:              logger,
		deletionProbability: 0.6, // def
		legacyFilter:        NewLegacyMessageFilter(bot, logger),
	}

	// conf
	if userIDStr := os.Getenv("TARGET_USER_ID"); userIDStr != "" {
		if userID, err := strconv.ParseInt(userIDStr, 10, 64); err == nil {
			service.targetUserID = userID
		} else {
			logger.Warn("invalid TARGET_USER_ID in environment", slog.String("value", userIDStr))
		}
	}

	if probStr := os.Getenv("DELETION_PROBABILITY"); probStr != "" {
		if prob, err := strconv.ParseFloat(probStr, 64); err == nil && prob >= 0 && prob <= 1 {
			service.deletionProbability = prob
		} else {
			logger.Warn("invalid DELETION_PROBABILITY in environment", slog.String("value", probStr))
		}
	}

	return service
}

func (s *MessageFilterService) HandleMessage(msg *tb.Message) error {
	if err := s.legacyFilter.HandleLegacyMessage(msg); err != nil {
		return err
	}

	if s.targetUserID == 0 || msg.Sender.ID != s.targetUserID {
		return nil
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	if r.Float64() < s.deletionProbability {
		if err := s.bot.Delete(msg); err != nil {
			s.logger.Error("failed to delete message",
				slog.Int64("user_id", msg.Sender.ID),
				slog.Int64("message_id", int64(msg.ID)),
				slog.String("error", err.Error()))
			return err
		}

		s.logger.Info("message deleted",
			slog.Int64("user_id", msg.Sender.ID),
			slog.Int64("message_id", int64(msg.ID)))
	}

	return nil
}
