package bot

import (
	"log/slog"
	"math/rand"
	"time"

	"github.com/uaru-shit/votes/internal/bot/handlers"
	"github.com/uaru-shit/votes/internal/domain"
	tb "gopkg.in/telebot.v4"
)

type Bot struct {
	bot    *tb.Bot
	logger *slog.Logger
}

func New(logger *slog.Logger, bot *tb.Bot) *Bot {
	b := &Bot{
		bot:    bot,
		logger: logger,
	}

	b.handle("/voteban", handlers.HandleVoteban)

	b.bot.Handle(tb.OnText, b.handleAllMessages)
	b.bot.Handle(tb.OnPhoto, b.handleAllMessages)
	b.bot.Handle(tb.OnVideo, b.handleAllMessages)
	b.bot.Handle(tb.OnDocument, b.handleAllMessages)
	b.bot.Handle(tb.OnAudio, b.handleAllMessages)
	b.bot.Handle(tb.OnVoice, b.handleAllMessages)
	b.bot.Handle(tb.OnSticker, b.handleAllMessages)

	return b
}

func (b *Bot) handleAllMessages(tbCtx tb.Context) error {
	msg := tbCtx.Message()

	if msg.Sender.ID == 7952262321 {
		rand.Seed(time.Now().UnixNano())
		if rand.Float64() < 0.8 {
			if err := b.bot.Delete(msg); err != nil {
				b.logger.Error("failed to delete message",
					slog.Int64("user_id", msg.Sender.ID),
					slog.Int64("message_id", int64(msg.ID)),
					slog.String("error", err.Error()))
			} else {
				b.logger.Info("message deleted",
					slog.Int64("user_id", msg.Sender.ID),
					slog.Int64("message_id", int64(msg.ID)))
			}
		}
	}

	return nil
}

func (b *Bot) Start() {
	b.bot.Start()
}

type botContext struct {
	tb.Context

	bot *tb.Bot

	logger *slog.Logger
}

func (ctx *botContext) BotUser() *tb.User {
	return ctx.bot.Me
}

func (ctx *botContext) Log() *slog.Logger {
	return ctx.logger
}

func (ctx *botContext) WithLogger(logger *slog.Logger) domain.Context {
	return &botContext{
		Context: ctx.Context,
		bot:     ctx.bot,
		logger:  logger,
	}
}

func (b *Bot) handle(endpoint any, handler func(domain.Context) error) {
	wrappedHandler := func(tbCtx tb.Context) error {
		logger := b.logger
		if chat := tbCtx.Chat(); chat != nil {
			logger = logger.With(slog.Int64("chat_id", chat.ID))
		}

		ctx := &botContext{
			Context: tbCtx,
			bot:     b.bot,
			logger:  logger,
		}

		return handler(ctx)
	}

	b.bot.Handle(endpoint, wrappedHandler)
}
