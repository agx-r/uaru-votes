package main

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
	tb "gopkg.in/telebot.v4"
)

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
			return zeroLevel, fmt.Errorf(`provided level string is not one of debug/info/warn/error neither a valid int`)
		}
		return slog.Level(levelInt), nil
	}
}

func errorAttr(err error) slog.Attr {
	return slog.String("err", err.Error())
}

type ErrorHandler struct {
	logger *slog.Logger
}

func (h *ErrorHandler) handleError(err error, ctx tb.Context) {
	if err != nil {
		h.logger.Error("error from bot", errorAttr(err))
	}
}

type Model struct {
	logger *slog.Logger
}

func main() {
	_ = godotenv.Load()

	var levelVar slog.LevelVar

	levelEnv, isSet := os.LookupEnv("VOTEBAN_LOG_LEVEL")

	level := slog.LevelInfo
	if isSet {
		levelParsed, err := parseLogLevel(levelEnv)
		if err != nil {
			fmt.Println("log level environment variable provided, but cannot parse it:", err)
			os.Exit(1)
		}

		level = levelParsed
	}

	levelVar.Set(level)

	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: &levelVar,
	}))

	errorHandler := ErrorHandler{
		logger: log,
	}

	token, isSet := os.LookupEnv("VOTEBAN_TG_TOKEN")
	if !isSet {
		log.Error("bot token variable not set")
		os.Exit(1)
	}

	bot, err := tb.NewBot(tb.Settings{
		Token:   token,
		OnError: errorHandler.handleError,
	})

	if err != nil {
		log.Error("failed to initialize bot:", errorAttr(err))
		os.Exit(1)
	}

	bot.Handle("/start", func(ctx tb.Context) error {
		return ctx.Reply("Приве")
	})

	bot.Handle("/voteban", func(ctx tb.Context) error {
		if !ctx.Message().FromGroup() {
			return ctx.Reply("В лс не баню сори")
		}

		if ctx.Message().ReplyTo == nil {
			return ctx.Reply("Ответь на сообщение кого забанить")
		}

		userToBan := ctx.Message().ReplyTo.Sender
		member, err := bot.ChatMemberOf(ctx.Chat(), userToBan)
		if err != nil {
			return err
		}

		admins, err := bot.AdminsOf(ctx.Chat())
		if err != nil {
			return err
		}

		amIAdmin := false
		for _, admin := range admins {
			if userToBan.ID == admin.User.ID {
				return ctx.Reply("ммм не")
			}
			if admin.CanRestrictMembers && admin.User.ID == bot.Me.ID {
				amIAdmin = true
			}
		}
		if !amIAdmin {
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
			return err
		}

		go func() {
			time.Sleep(time.Hour)

			poll, err := bot.StopPoll(msg)
			if err != nil {
				log.Error("failed to stop the poll", errorAttr(err))
				return
			}

			if poll.Options[0].VoterCount > poll.Options[1].VoterCount {
				if _, err := bot.Reply(msg, "Мут nyuuu"); err != nil {
					log.Error("failed to reply to poll", errorAttr(err))
				}

				member.RestrictedUntil = tb.Forever()
				member.CanSendMessages = false
				member.CanSendMedia = false
				member.CanSendPolls = false
				member.CanSendOther = false
				member.CanAddPreviews = false
				if err := bot.Restrict(ctx.Chat(), member); err != nil {
					log.Error("cannot mute user", errorAttr(err))
					if _, err := bot.Reply(msg, "Чота не могу замутить"); err != nil {
						log.Error("can't even cry", errorAttr(err))
					}
				}
			} else {
				if _, err := bot.Reply(msg, "Размучен."); err != nil {
					log.Error("failed to reply to poll", errorAttr(err))
				}

				member.RestrictedUntil = 0
				member.CanSendMessages = true
				member.CanSendMedia = true
				member.CanSendPolls = true
				member.CanSendOther = true
				member.CanAddPreviews = true
				if err := bot.Restrict(ctx.Chat(), member); err != nil {
					log.Error("cannot unmute user", errorAttr(err))
					if _, err := bot.Reply(msg, "Чота не могу размутить"); err != nil {
						log.Error("can't even cry", errorAttr(err))
					}
				}
			}
		}()

		return nil
	})

	bot.Start()
}