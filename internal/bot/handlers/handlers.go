package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/uaru-shit/votes/internal/domain"
	"github.com/uaru-shit/votes/pkg/utils"
	tb "gopkg.in/telebot.v4"
)

func generatePollID() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func createPoll(ctx domain.Context, pollType domain.PollType, user *tb.User, member *tb.ChatMember, question string, options []string) error {
	bot := ctx.BotAPI()
	pollID := generatePollID()
	
	// Convert string options to PollOption structs
	pollOptions := make([]tb.PollOption, len(options))
	for i, option := range options {
		pollOptions[i] = tb.PollOption{Text: option}
	}
	
	msg, err := bot.Reply(ctx.Message(), &tb.Poll{
		Question:  question,
		Anonymous: false,
		Options:   pollOptions,
	})
	if err != nil {
		return fmt.Errorf("failed to send poll: %w", err)
	}

	memberData, err := utils.SerializeMember(member)
	if err != nil {
		return fmt.Errorf("failed to serialize member: %w", err)
	}

	pollDuration := getPollDuration(ctx)

	const (
		minDuration = 30 * time.Second
		maxDuration = 24 * time.Hour
	)

	if pollDuration < minDuration {
		ctx.Log().Warn("poll duration too short, using minimum", 
			slog.String("provided", pollDuration.String()),
			slog.String("minimum", minDuration.String()))
		pollDuration = minDuration
	} else if pollDuration > maxDuration {
		ctx.Log().Warn("poll duration too long, using maximum", 
			slog.String("provided", pollDuration.String()),
			slog.String("maximum", maxDuration.String()))
		pollDuration = maxDuration
	}
	activePoll := &domain.ActivePoll{
		ID:         pollID,
		Type:       pollType,
		ChatID:     ctx.Chat().ID,
		MessageID:  msg.ID,
		UserID:     user.ID,
		CreatedAt:  time.Now(),
		ExpiresAt:  time.Now().Add(pollDuration),
		MemberData: memberData,
	}

	if err := ctx.PollStorage().SavePoll(activePoll); err != nil {
		return fmt.Errorf("failed to save poll: %w", err)
	}

	ctx.StartPollMonitoring(activePoll)
	return nil
}

func getPollDuration(ctx domain.Context) time.Duration {
	const defaultDuration = 30 * time.Minute

	durationStr := os.Getenv("VOTEBAN_POLL_DURATION_SECONDS")
	if durationStr == "" {
		ctx.Log().Debug("using default poll duration", 
			slog.String("duration", defaultDuration.String()))
		return defaultDuration
	}

	duration, err := strconv.Atoi(durationStr)
	if err != nil || duration <= 0 {
		ctx.Log().Warn("invalid poll duration in environment", 
			slog.String("value", durationStr),
			slog.String("error", err.Error()))
		return defaultDuration
	}

	pollDuration := time.Duration(duration) * time.Second
	ctx.Log().Info("using custom poll duration", 
		slog.Int("seconds", duration),
		slog.String("duration", pollDuration.String()))
	
	return pollDuration
}



func validatePollRequest(ctx domain.Context, user *tb.User) (*tb.ChatMember, error) {
	bot := ctx.BotAPI()

	if !ctx.Message().FromGroup() {
		return nil, fmt.Errorf("команда работает только в группах")
	}

	member, err := bot.ChatMemberOf(ctx.Chat(), user)
	if err != nil {
		return nil, fmt.Errorf("failed to get member: %w", err)
	}

	admins, err := bot.AdminsOf(ctx.Chat())
	if err != nil {
		return nil, fmt.Errorf("failed to get admins: %w", err)
	}

	if utils.IsAdmin(user.ID, admins) {
		return nil, fmt.Errorf("нельзя голосовать против администраторов")
	}

	if !utils.BotCanMute(ctx.BotUser().ID, admins) {
		return nil, fmt.Errorf("бот должен быть администратором")
	}

	return member, nil
}

func validateAdminAccess(ctx domain.Context) error {
	if os.Getenv("ADMINS_ONLY") == "true" {
		bot := ctx.BotAPI()
		admins, err := bot.AdminsOf(ctx.Chat())
		if err != nil {
			return fmt.Errorf("failed to get admins: %w", err)
		}

		if !utils.IsAdmin(ctx.Message().Sender.ID, admins) {
			return fmt.Errorf("команда доступна только администраторам")
		}
	}
	return nil
}

func HandleVoteban(ctx domain.Context) error {
	if err := validateAdminAccess(ctx); err != nil {
		return ctx.Reply(err.Error())
	}

	if ctx.Message().ReplyTo == nil {
		return ctx.Reply("ответь на сообщение")
	}
	userToBan := ctx.Message().ReplyTo.Sender

	member, err := validatePollRequest(ctx, userToBan)
	if err != nil {
		return ctx.Reply(err.Error())
	}

	return createPoll(ctx, domain.PollTypeBan, userToBan, member, "Банить?", []string{"Да", "Нет"})
}

func HandleVoteUnban(ctx domain.Context) error {
	if err := validateAdminAccess(ctx); err != nil {
		return ctx.Reply(err.Error())
	}

	if ctx.Message().ReplyTo == nil {
		return ctx.Reply("ответь на сообщение")
	}
	userToUnban := ctx.Message().ReplyTo.Sender

	member, err := validatePollRequest(ctx, userToUnban)
	if err != nil {
		return ctx.Reply(err.Error())
	}

	return createPoll(ctx, domain.PollTypeUnban, userToUnban, member, "Разбанить?", []string{"Да", "Нет"})
}

func HandleVoteGifs(ctx domain.Context) error {
	if err := validateAdminAccess(ctx); err != nil {
		return ctx.Reply(err.Error())
	}

	if ctx.Message().ReplyTo == nil {
		return ctx.Reply("ответь на сообщение")
	}
	user := ctx.Message().ReplyTo.Sender

	member, err := validatePollRequest(ctx, user)
	if err != nil {
		return ctx.Reply(err.Error())
	}

	return createPoll(ctx, domain.PollTypeGifs, user, member, "Стикеры/гифки этому челу:", []string{"Запретить", "Разрешить"})
}

func HandleVoteMedia(ctx domain.Context) error {
	if err := validateAdminAccess(ctx); err != nil {
		return ctx.Reply(err.Error())
	}

	if ctx.Message().ReplyTo == nil {
		return ctx.Reply("ответь на сообщение")
	}
	user := ctx.Message().ReplyTo.Sender

	member, err := validatePollRequest(ctx, user)
	if err != nil {
		return ctx.Reply(err.Error())
	}

	return createPoll(ctx, domain.PollTypeMedia, user, member, "Медиа этому челу:", []string{"Запретить", "Разрешить"})
}

func HandleInstaban(ctx domain.Context) error {
	if err := validateAdminAccess(ctx); err != nil {
		return ctx.Reply(err.Error())
	}

	if ctx.Message().ReplyTo == nil {
		return ctx.Reply("ответь на сообщение")
	}
	userToBan := ctx.Message().ReplyTo.Sender

	if !ctx.Message().FromGroup() {
		return ctx.Reply("команда работает только в группах")
	}

	bot := ctx.BotAPI()
	member, err := bot.ChatMemberOf(ctx.Chat(), userToBan)
	if err != nil {
		return ctx.Reply("не могу получить данные пользователя")
	}

	admins, err := bot.AdminsOf(ctx.Chat())
	if err != nil {
		return ctx.Reply("не могу получить список администраторов")
	}

	if utils.IsAdmin(userToBan.ID, admins) {
		return ctx.Reply("нельзя банить администраторов")
	}

	if !utils.BotCanMute(ctx.BotUser().ID, admins) {
		return ctx.Reply("бот должен быть администратором")
	}

	// Мгновенный бан без опроса
	if err := bot.Ban(ctx.Chat(), &tb.ChatMember{User: userToBan}); err != nil {
		ctx.Log().Error("failed to ban user", 
			slog.Int64("user_id", userToBan.ID),
			slog.String("error", err.Error()))
		return ctx.Reply("не удалось забанить пользователя")
	}

	ctx.Log().Info("user banned instantly", 
		slog.Int64("user_id", userToBan.ID),
		slog.String("username", userToBan.Username),
		slog.Int64("admin_id", ctx.Message().Sender.ID))

	return ctx.Reply(fmt.Sprintf("Пользователь @%s забанен мгновенно", userToBan.Username))
}

func HandleHelp(ctx domain.Context) error {
	helpText := `<b>COMMANDS</b>

<b>Ban/Unban:</b>
/ban - Start vote to ban user
/unban - Start vote to unban user
/instaban - Instantly ban user (admin only)

<b>Permissions:</b>
/gif - Start vote to restrict gifs/stickers
/media - Start vote to restrict media

<b>Usage:</b> Reply to any message with a command to start voting.`

	_, err := ctx.BotAPI().Reply(ctx.Message(), helpText, &tb.SendOptions{
		ParseMode: tb.ModeHTML,
	})
	return err
}


