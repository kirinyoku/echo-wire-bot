package main

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"syscall"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jmoiron/sqlx"
	"github.com/kirinyoku/echo-wire-bot/internal/bot"
	"github.com/kirinyoku/echo-wire-bot/internal/bot/middleware"
	"github.com/kirinyoku/echo-wire-bot/internal/botkit"
	"github.com/kirinyoku/echo-wire-bot/internal/config"
	"github.com/kirinyoku/echo-wire-bot/internal/fetcher"
	"github.com/kirinyoku/echo-wire-bot/internal/notifier"
	"github.com/kirinyoku/echo-wire-bot/internal/storage"
	"github.com/kirinyoku/echo-wire-bot/internal/summary"
	_ "github.com/lib/pq"
)

func main() {
	botAPI, err := tgbotapi.NewBotAPI(config.Get().TelegramBotToken)
	if err != nil {
		log.Printf("failed to initialize bot: %v", err)
		return
	}

	db, err := sqlx.Connect("postgres", config.Get().DatabaseDSN)
	if err != nil {
		log.Printf("failed to connect to database: %v", err)
		return
	}

	defer db.Close()

	var (
		articleStorage = storage.NewArticleStorage(db)
		sourceStorage  = storage.NewSourceStorage(db)
		summarizer     = summary.NewOpenAISummarizer(config.Get().OpenAIKey, config.Get().OpenAIModel, config.Get().OpenAIPrompt)
		fetcher        = fetcher.New(articleStorage, sourceStorage, config.Get().FetchInterval, config.Get().FilterKeywords)
		notifier       = notifier.New(
			articleStorage,
			summarizer, botAPI,
			config.Get().NotificationInterval,
			config.Get().FetchInterval,
			config.Get().TelegramChannelID,
		)
	)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	newsBot := botkit.New(botAPI)
	newsBot.RegisterCommand("addsource", middleware.AdminsOnly(config.Get().TelegramChannelID, bot.ViewCmdAddSource(sourceStorage)))
	newsBot.RegisterCommand("deletesource", middleware.AdminsOnly(config.Get().TelegramChannelID, bot.ViewCmdDeleteSource(sourceStorage)))
	newsBot.RegisterCommand("getsource", middleware.AdminsOnly(config.Get().TelegramChannelID, bot.ViewCmdGetSource(sourceStorage)))
	newsBot.RegisterCommand("listsources", middleware.AdminsOnly(config.Get().TelegramChannelID, bot.ViewCmdListSources(sourceStorage)))

	go func(ctx context.Context) {
		if err := fetcher.Run(ctx); err != nil {
			if !errors.Is(err, context.Canceled) {
				log.Printf("failed to run fetcher: %v", err)
				return
			}

			log.Printf("fetcher stopped")
		}
	}(ctx)

	go func(ctx context.Context) {
		if err := notifier.Start(ctx); err != nil {
			if !errors.Is(err, context.Canceled) {
				log.Printf("failed to run notifier: %v", err)
				return
			}

			log.Printf("notifier stopped")
		}
	}(ctx)

	if err := newsBot.Start(ctx); err != nil {
		log.Printf("failed to start bot: %v", err)
		return
	}
}
