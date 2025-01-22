package bot

import (
	"context"
	"fmt"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/kirinyoku/echo-wire-bot/internal/botkit"
	"github.com/kirinyoku/echo-wire-bot/internal/botkit/markup"
	"github.com/kirinyoku/echo-wire-bot/internal/models"
)

// SourceProvider is an interface for retrieving a source from persistent storage.
// It defines the SourceByID method, which fetches a source by its ID.
type SourceProvider interface {
	SourceByID(ctx context.Context, id int64) (models.Source, error)
}

// ViewCmdGetSource creates a bot command handler for retrieving source details by ID.
// It parses the source ID from the command arguments, fetches the source from storage,
// and sends a formatted message with the source details to the user.
func ViewCmdGetSource(provider SourceProvider) botkit.ViewFunc {
	return func(ctx context.Context, bot *tgbotapi.BotAPI, update tgbotapi.Update) error {
		idStr := update.Message.CommandArguments()

		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			return err
		}

		source, err := provider.SourceByID(ctx, id)
		if err != nil {
			return err
		}

		reply := tgbotapi.NewMessage(update.Message.Chat.ID, formatSource(source))
		reply.ParseMode = parseModeMarkdownV2

		if _, err := bot.Send(reply); err != nil {
			return err
		}

		return nil
	}
}

// formatSource formats the details of a source into a Markdown-compatible string.
// It escapes special Markdown characters to ensure proper rendering in the message.
func formatSource(source models.Source) string {
	return fmt.Sprintf(
		"Name: *%s*\nID: `%d`\nFeed URL: %s",
		markup.EscapeForMarkdown(source.Name),
		source.ID,
		markup.EscapeForMarkdown(source.URL),
	)
}
