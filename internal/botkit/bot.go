package botkit

import (
	"context"
	"fmt"
	"log"
	"runtime/debug"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Bot represents a wrapper around the Telegram Bot API with command handling functionality.
type Bot struct {
	api      *tgbotapi.BotAPI
	cmdViems map[string]ViewFunc
}

// ViewFunc defines a function type for handling Telegram updates.
// Parameters:
//   - ctx: The context for managing request deadlines and cancellations.
//   - bot: The Telegram Bot API instance.
//   - update: The update received from Telegram.
type ViewFunc func(ctx context.Context, bot *tgbotapi.BotAPI, update tgbotapi.Update) error

// New creates a new Bot instance with the provided Telegram Bot API.
func New(api *tgbotapi.BotAPI) *Bot {
	return &Bot{
		api: api,
	}
}

// RegisterCommand registers a command with its corresponding view function in the bot.
func (b *Bot) RegisterCommand(name string, view ViewFunc) {
	if b.cmdViems == nil {
		b.cmdViems = make(map[string]ViewFunc)
	}

	b.cmdViems[name] = view
}

// handleUpdate processes a single Telegram update, invoking the appropriate view function.
func (b *Bot) handleUpdate(ctx context.Context, update tgbotapi.Update) {
	const op = "bot.handleUpdate"

	defer func() {
		if p := recover(); p != nil {
			log.Fatalf("panic recovered: %v\n%s", p, string(debug.Stack()))
		}
	}()

	if update.Message == nil || !update.Message.IsCommand() {
		return
	}

	var view ViewFunc

	cmd := update.Message.Command()

	cmdView, ok := b.cmdViems[cmd]
	if !ok {
		return
	}

	view = cmdView

	if err := view(ctx, b.api, update); err != nil {
		log.Printf("%s: %v", op, err)
	}
}

// Start begins listening for updates from Telegram and processes them.
func (b *Bot) Start(ctx context.Context) error {
	const op = "bot.Start"

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.api.GetUpdatesChan(u)

	for {
		select {
		case update := <-updates:
			updateCtx, updateCancel := context.WithTimeout(ctx, 5*time.Second)
			b.handleUpdate(updateCtx, update)
			updateCancel()
		case <-ctx.Done():
			return fmt.Errorf("%s: context done", op)
		}
	}
}
