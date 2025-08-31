// Simple bot to ban group members based on poll results
package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/joho/godotenv"
	"github.com/uaru-shit/votes/internal/bot"
	"github.com/uaru-shit/votes/pkg/utils"
	tb "gopkg.in/telebot.v4"
)

func main() {
	_ = godotenv.Load()

	var levelVar slog.LevelVar

	levelEnv, isSet := os.LookupEnv("VOTEBAN_LOG_LEVEL")

	level := slog.LevelInfo

	if isSet {
		levelParsed, err := utils.ParseLogLevel(levelEnv)
		if err != nil {
			fmt.Fprintf(os.Stderr, "log level environment variable set, but cannot parse it: %v\n", err)
			os.Exit(1)
		}

		level = levelParsed
	}

	levelVar.Set(level)

	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: &levelVar,
	}))

	eHandler := utils.NewErrorHandler(log)

	token, isSet := os.LookupEnv("VOTEBAN_TG_TOKEN")
	if !isSet {
		log.Error("bot token variable not set")
		os.Exit(1)
	}

	tbBot, err := tb.NewBot(tb.Settings{
		Token:   token,
		OnError: eHandler.HandleError,
	})
	if err != nil {
		log.Error("failed to initialize bot:", utils.ErrorAttr(err))
		os.Exit(1)
	}

	// create poll storage
	pollStorage, err := utils.NewFilePollStorage("data/active_polls.json")
	if err != nil {
		log.Error("failed to create poll storage:", utils.ErrorAttr(err))
		os.Exit(1)
	}

	b := bot.New(log, tbBot, pollStorage)

	b.Start()
}
