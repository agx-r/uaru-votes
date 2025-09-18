package bot

import (
	"log/slog"

	"github.com/uaru-shit/votes/internal/bot/handlers"
	"github.com/uaru-shit/votes/internal/domain"
	"github.com/uaru-shit/votes/internal/services"
	tb "gopkg.in/telebot.v4"
)

type Bot struct {
	bot           *tb.Bot
	logger        *slog.Logger
	pollStorage   domain.PollStorage
	pollMonitor   *services.PollMonitorService
	messageFilter *services.MessageFilterService
}

func New(logger *slog.Logger, bot *tb.Bot, pollStorage domain.PollStorage) *Bot {
	permissionService := services.NewPermissionService(bot, logger)
	pollProcessor := services.NewPollProcessorService(bot, logger, permissionService)
	pollMonitor := services.NewPollMonitorService(bot, logger, pollStorage, pollProcessor)
	messageFilter := services.NewMessageFilterService(bot, logger)

	b := &Bot{
		bot:           bot,
		logger:        logger,
		pollStorage:   pollStorage,
		pollMonitor:   pollMonitor,
		messageFilter: messageFilter,
	}

	b.setupHandlers()
	go pollMonitor.RestoreActivePolls()

	return b
}

func (b *Bot) setupHandlers() {
	b.handle("/voteban", handlers.HandleVoteban)
	b.handle("/vote", handlers.HandleVoteban)
	b.handle("/ban", handlers.HandleVoteban)

	b.handle("/voteunban", handlers.HandleVoteUnban)
	b.handle("/unban", handlers.HandleVoteUnban)

	b.handle("/instaban", handlers.HandleInstaban)

	b.handle("/votegif", handlers.HandleVoteGifs)
	b.handle("/gif", handlers.HandleVoteGifs)

	b.handle("/votemedia", handlers.HandleVoteMedia)
	b.handle("/media", handlers.HandleVoteMedia)

	b.handle("/help", handlers.HandleHelp)

	b.bot.Handle(tb.OnText, b.handleAllMessages)
	b.bot.Handle(tb.OnPhoto, b.handleAllMessages)
	b.bot.Handle(tb.OnVideo, b.handleAllMessages)
	b.bot.Handle(tb.OnDocument, b.handleAllMessages)
	b.bot.Handle(tb.OnAudio, b.handleAllMessages)
	b.bot.Handle(tb.OnVoice, b.handleAllMessages)
	b.bot.Handle(tb.OnSticker, b.handleAllMessages)
}

func (b *Bot) handleAllMessages(tbCtx tb.Context) error {
	return b.messageFilter.HandleMessage(tbCtx.Message())
}

func (b *Bot) Start() {
	b.bot.Start()
}

type botContext struct {
	tb.Context

	bot         *Bot
	logger      *slog.Logger
	pollStorage domain.PollStorage
}

func (ctx *botContext) BotUser() *tb.User {
	return ctx.bot.bot.Me
}

func (ctx *botContext) BotAPI() tb.API {
	return ctx.bot.bot
}

func (ctx *botContext) Log() *slog.Logger {
	return ctx.logger
}

func (ctx *botContext) WithLogger(logger *slog.Logger) domain.Context {
	return &botContext{
		Context: ctx.Context,
		bot:     ctx.bot,
		logger:  logger,
		pollStorage: ctx.pollStorage,
	}
}

func (ctx *botContext) PollStorage() domain.PollStorage {
	return ctx.pollStorage
}

func (ctx *botContext) StartPollMonitoring(poll *domain.ActivePoll) {
	ctx.bot.pollMonitor.StartPollMonitoring(poll)
}

func (b *Bot) handle(endpoint any, handler func(domain.Context) error) {
	wrappedHandler := func(tbCtx tb.Context) error {
		logger := b.logger
		if chat := tbCtx.Chat(); chat != nil {
			logger = logger.With(slog.Int64("chat_id", chat.ID))
		}

		ctx := &botContext{
			Context:     tbCtx,
			bot:         b,
			logger:      logger,
			pollStorage: b.pollStorage,
		}

		return handler(ctx)
	}

	b.bot.Handle(endpoint, wrappedHandler)
}
