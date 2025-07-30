// Simple bot to ban group members based on poll results
package main

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
	tb "gopkg.in/telebot.v4"
)

var errInvalidLogLevel = errors.New("provided level string is not one of debug/info/warn/error neither a valid int")

func parseLogLevel(level string) (slog.Level, error) {
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

			return zeroLevel, errInvalidLogLevel
		}

		return slog.Level(levelInt), nil
	}
}

func errorAttr(err error) slog.Attr {
	return slog.String("err", err.Error())
}

type errorHandler struct {
	logger *slog.Logger
}

func (h *errorHandler) handleError(err error, _ tb.Context) {
	if err != nil {
		h.logger.Error("error from bot", errorAttr(err))
	}
}

func isAdmin(userID int64, admins []tb.ChatMember) bool {
	for _, admin := range admins {
		if admin.User.ID == userID {
			return true
		}
	}

	return false
}

func botCanMute(botID int64, admins []tb.ChatMember) bool {
	for _, admin := range admins {
		if admin.User.ID == botID && admin.CanRestrictMembers {
			return true
		}
	}

	return false
}

func handlePollResult(bot *tb.Bot, log *slog.Logger, msg *tb.Message, member *tb.ChatMember, ctx tb.Context) {
	time.Sleep(time.Hour)

	poll, err := bot.StopPoll(msg)
	if err != nil {
		log.Error("failed to stop the poll", errorAttr(err))

		return
	}

	shouldMute := poll.Options[0].VoterCount > poll.Options[1].VoterCount

	if _, err := bot.Reply(msg, map[bool]string{
		true:  "Мут nyuuu",
		false: "Размучен.",
	}[shouldMute]); err != nil {
		log.Error("failed to reply to poll", errorAttr(err))
	}

	perm := !shouldMute // value of permissions is opposite of shouldMute

	member.RestrictedUntil = map[bool]int64{
		true:  tb.Forever(), // tb.Forever() is int64
		false: 0,
	}[shouldMute]
	member.CanSendMessages = perm
	member.CanSendMedia = perm
	member.CanSendPolls = perm
	member.CanSendOther = perm
	member.CanAddPreviews = perm

	if err := bot.Restrict(ctx.Chat(), member); err != nil {
		logWord := map[bool]string{true: "mute", false: "unmute"}[shouldMute]
		msgWord := map[bool]string{true: "замутить", false: "размутить"}[shouldMute]

		log.Error(fmt.Sprintf("cannot %s user", logWord), errorAttr(err))

		if _, err := bot.Reply(msg, "Чота не могу "+msgWord); err != nil {
			log.Error("can't even cry", errorAttr(err))
		}
	}
}

func handleVoteban(bot *tb.Bot, log *slog.Logger) func(tb.Context) error {
	return func(ctx tb.Context) error {
		if !ctx.Message().FromGroup() {
			return ctx.Reply("В лс не баню сори")
		}

		if ctx.Message().ReplyTo == nil {
			return ctx.Reply("Ответь на сообщение кого забанить")
		}

		userToBan := ctx.Message().ReplyTo.Sender
		member, err := bot.ChatMemberOf(ctx.Chat(), userToBan)

		if err != nil {
			return fmt.Errorf("failed to get member: %w", err)
		}

		admins, err := bot.AdminsOf(ctx.Chat())
		if err != nil {
			return fmt.Errorf("failed to get admins: %w", err)
		}

		if isAdmin(userToBan.ID, admins) {
			return ctx.Reply("ммм не")
		}

		if !botCanMute(bot.Me.ID, admins) {
			return ctx.Reply("Админом меня сделай, олух")
		}

		msg, err := bot.Send(ctx.Chat(), &tb.Poll{
			Question:  "Забанить или разбанить?",
			Anonymous: false,
			Options: []tb.PollOption{
				{Text: "Мут"},
				{Text: "Размут"},
			},
		})
		if err != nil {
			return fmt.Errorf("failed to send poll: %w", err)
		}

		go handlePollResult(bot, log, msg, member, ctx)

		return nil
	}
}

func main() {
	_ = godotenv.Load()

	var levelVar slog.LevelVar

	levelEnv, isSet := os.LookupEnv("VOTEBAN_LOG_LEVEL")

	level := slog.LevelInfo

	if isSet {
		levelParsed, err := parseLogLevel(levelEnv)
		if err != nil {
			fmt.Fprintf(os.Stderr, "log level environment variable provided, but cannot parse it: %v\n", err)
			os.Exit(1)
		}

		level = levelParsed
	}

	levelVar.Set(level)

	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: &levelVar,
	}))

	eHandler := errorHandler{
		logger: log,
	}

	token, isSet := os.LookupEnv("VOTEBAN_TG_TOKEN")
	if !isSet {
		log.Error("bot token variable not set")
		os.Exit(1)
	}

	bot, err := tb.NewBot(tb.Settings{
		Token:   token,
		OnError: eHandler.handleError,
	})
	if err != nil {
		log.Error("failed to initialize bot:", errorAttr(err))
		os.Exit(1)
	}

	bot.Handle("/start", func(ctx tb.Context) error { return ctx.Reply("Приве") })

	bot.Handle("/voteban", handleVoteban(bot, log))

	bot.Start()
}
