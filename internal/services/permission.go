package services

import (
	"fmt"
	"log/slog"

	tb "gopkg.in/telebot.v4"
)

type PermissionService struct {
	bot    tb.API
	logger *slog.Logger
}

func NewPermissionService(bot tb.API, logger *slog.Logger) *PermissionService {
	return &PermissionService{
		bot:    bot,
		logger: logger,
	}
}

func (s *PermissionService) UpdatePermission(msg *tb.Message, member *tb.ChatMember, permission string, value bool, errorMsg, successMsg string) error {
	currentMember, err := s.bot.ChatMemberOf(msg.Chat, member.User)
	if err != nil {
		s.logger.Error("cannot get current member data", slog.String("error", err.Error()))
		_, replyErr := s.bot.Reply(msg, "Чота не могу получить данные пользователя")
		if replyErr != nil {
			s.logger.Error("failed to send error message", slog.String("error", replyErr.Error()))
		}
		return err
	}

	currentMember.Independent = true

	switch permission {
	case "CanSendOther":
		currentMember.CanSendOther = value
	case "CanSendMedia":
		currentMember.CanSendPhotos = value
		currentMember.CanSendVideos = value
		currentMember.CanSendDocuments = value
		currentMember.CanSendAudios = value
		currentMember.CanSendVoiceNotes = value
		currentMember.CanSendVideoNotes = value
	default:
		return fmt.Errorf("unknown permission: %s", permission)
	}

	s.logger.Info("updating member permissions", 
		slog.String("permission", permission),
		slog.Bool("value", value),
		slog.Int64("user_id", currentMember.User.ID))

	if err := s.bot.Restrict(msg.Chat, currentMember); err != nil {
		s.logger.Error("cannot update permission", 
			slog.String("permission", permission),
			slog.Bool("value", value),
			slog.String("error", err.Error()))
		_, replyErr := s.bot.Reply(msg, errorMsg)
		if replyErr != nil {
			s.logger.Error("failed to send error message", slog.String("error", replyErr.Error()))
		}
		return err
	}

	_, err = s.bot.Reply(msg, successMsg)
	if err != nil {
		s.logger.Error("failed to reply to poll", slog.String("error", err.Error()))
	}
	return err
}
