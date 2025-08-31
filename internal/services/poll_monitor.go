package services

import (
	"context"
	"log/slog"
	"time"

	"github.com/uaru-shit/votes/internal/domain"
	"github.com/uaru-shit/votes/pkg/utils"
	tb "gopkg.in/telebot.v4"
)

type PollMonitorService struct {
	bot         tb.API
	logger      *slog.Logger
	pollStorage domain.PollStorage
	processor   *PollProcessorService
}

func NewPollMonitorService(bot tb.API, logger *slog.Logger, pollStorage domain.PollStorage, processor *PollProcessorService) *PollMonitorService {
	return &PollMonitorService{
		bot:         bot,
		logger:      logger,
		pollStorage: pollStorage,
		processor:   processor,
	}
}

func (s *PollMonitorService) RestoreActivePolls() {
	polls, err := s.pollStorage.GetPolls()
	if err != nil {
		s.logger.Error("failed to load active polls", slog.String("error", err.Error()))
		return
	}

	s.logger.Info("restoring active polls", slog.Int("count", len(polls)))

	for _, poll := range polls {
		go s.monitorPoll(context.Background(), poll)
	}
}

func (s *PollMonitorService) StartPollMonitoring(poll *domain.ActivePoll) {
	go s.monitorPoll(context.Background(), poll)
}

func (s *PollMonitorService) monitorPoll(ctx context.Context, poll *domain.ActivePoll) {
	s.logger.Info("starting poll monitoring", 
		slog.String("poll_id", poll.ID),
		slog.String("type", string(poll.Type)),
		slog.Time("expires_at", poll.ExpiresAt))
	
	timeUntilExpiration := time.Until(poll.ExpiresAt)
	if timeUntilExpiration <= 0 {
		s.logger.Info("poll already expired, processing immediately", 
			slog.String("poll_id", poll.ID))
		s.processPoll(poll)
		return
	}

	s.logger.Info("waiting for poll to expire", 
		slog.String("poll_id", poll.ID),
		slog.String("duration", timeUntilExpiration.String()))
	
	timer := time.NewTimer(timeUntilExpiration)
	defer timer.Stop()

	select {
	case <-timer.C:
		s.logger.Info("poll expired, processing", slog.String("poll_id", poll.ID))
		s.processPoll(poll)
	case <-ctx.Done():
		s.logger.Info("poll monitoring cancelled", slog.String("poll_id", poll.ID))
		return
	}
}

func (s *PollMonitorService) processPoll(poll *domain.ActivePoll) {
	chat := &tb.Chat{ID: poll.ChatID}
	msg := &tb.Message{ID: poll.MessageID, Chat: chat}

	member, err := utils.DeserializeMember(poll.MemberData)
	if err != nil {
		s.logger.Error("failed to deserialize member data", 
			slog.String("poll_id", poll.ID),
			slog.String("error", err.Error()))
		return
	}

	if err := s.processor.ProcessExpiredPoll(poll.Type, msg, member); err != nil {
		s.logger.Error("failed to process expired poll",
			slog.String("poll_id", poll.ID),
			slog.String("error", err.Error()))
		return
	}

	if err := s.pollStorage.DeletePoll(poll.ID); err != nil {
		s.logger.Error("failed to delete poll from storage", 
			slog.String("poll_id", poll.ID),
			slog.String("error", err.Error()))
	}
}
