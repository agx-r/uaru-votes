package services

import (
	"fmt"
	"log/slog"

	"github.com/uaru-shit/votes/internal/domain"
	tb "gopkg.in/telebot.v4"
)

type PollProcessorService struct {
	bot    tb.API
	logger *slog.Logger
	perms  *PermissionService
}

func NewPollProcessorService(bot tb.API, logger *slog.Logger, perms *PermissionService) *PollProcessorService {
	return &PollProcessorService{
		bot:    bot,
		logger: logger,
		perms:  perms,
	}
}

func (s *PollProcessorService) ProcessExpiredPoll(pollType domain.PollType, msg *tb.Message, member *tb.ChatMember) error {
	poll, err := s.bot.StopPoll(msg)
	if err != nil {
		return fmt.Errorf("failed to stop poll: %w", err)
	}

	// for polls of type "Запретить/Разрешить"    the first option is   "Запретить"
	// for polls of type         "Да/Нет"         the first option is      "Да"
	shouldRestrict := poll.Options[0].VoterCount > poll.Options[1].VoterCount

	switch pollType {
	case domain.PollTypeBan:
		return s.processBanResult(msg, member, shouldRestrict)
	case domain.PollTypeUnban:
		return s.processUnbanResult(msg, member, shouldRestrict)
	case domain.PollTypeGifs:
		return s.processGifsResult(msg, member, shouldRestrict)
	case domain.PollTypeMedia:
		return s.processMediaResult(msg, member, shouldRestrict)
	default:
		return fmt.Errorf("unknown poll type: %s", pollType)
	}
}

func (s *PollProcessorService) processBanResult(msg *tb.Message, member *tb.ChatMember, shouldBan bool) error {
	if shouldBan {
		return s.handleBan(msg, member)
	}
	return s.handleUnban(msg, member)
}

func (s *PollProcessorService) processUnbanResult(msg *tb.Message, member *tb.ChatMember, shouldUnban bool) error {
	if shouldUnban {
		return s.handleUnban(msg, member)
	}
	return s.handleBan(msg, member)
}

func (s *PollProcessorService) processGifsResult(msg *tb.Message, member *tb.ChatMember, shouldMute bool) error {
	if shouldMute {
		return s.perms.UpdatePermission(msg, member, "CanSendOther", false, 
			"Чота не могу отключить стикеры", "-брейнрот")
	}
	return s.perms.UpdatePermission(msg, member, "CanSendOther", true, 
		"Чота не могу включить стикеры", "Брейнрот снова доступен")
}

func (s *PollProcessorService) processMediaResult(msg *tb.Message, member *tb.ChatMember, shouldMute bool) error {
	if shouldMute {
		return s.perms.UpdatePermission(msg, member, "CanSendMedia", false, 
			"Чота не могу отключить медиа", "Медиа заблокированы")
	}
	return s.perms.UpdatePermission(msg, member, "CanSendMedia", true, 
		"Чота не могу включить медиа", "Медиа снова доступны")
}

func (s *PollProcessorService) handleBan(msg *tb.Message, member *tb.ChatMember) error {
	if err := s.bot.Ban(msg.Chat, member); err != nil {
		s.logger.Error("cannot ban user", slog.String("error", err.Error()))
		_, replyErr := s.bot.Reply(msg, "Чота не могу забанить")
		if replyErr != nil {
			s.logger.Error("failed to send error message", slog.String("error", replyErr.Error()))
		}
		return err
	}

	_, err := s.bot.Reply(msg, "BAN B AN BAN BAN BANBANBANBAN BAN BANBANBA NB ANBANB ANBANB ANBANB ANBAN BAN BANBA NBNBANBANB AN BA NBA NBANBA NB ANB ANB AN BANB AN\n!!!!!!!!\n!!!!!!!!!!!!!!!!!!!!\n!!!!!!!!!!!!!!!!!\n!!!!!!!!!!!!!!!!!!\n!!!!!!!!\n!!!!!!!!!!!!")
	if err != nil {
		s.logger.Error("failed to reply to poll", slog.String("error", err.Error()))
	}
	return err
}

func (s *PollProcessorService) handleUnban(msg *tb.Message, member *tb.ChatMember) error {
	if err := s.bot.Unban(msg.Chat, member.User, true); err != nil {
		s.logger.Error("cannot unban user", slog.String("error", err.Error()))
		_, replyErr := s.bot.Reply(msg, "Чота не могу разбанить")
		if replyErr != nil {
			s.logger.Error("failed to send error message", slog.String("error", replyErr.Error()))
		}
		return err
	}

	_, err := s.bot.Reply(msg, "Разбанен")
	if err != nil {
		s.logger.Error("failed to reply to poll", slog.String("error", err.Error()))
	}
	return err
}
