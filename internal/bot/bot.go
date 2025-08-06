package bot

import (
	"log/slog"

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

	return b
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
