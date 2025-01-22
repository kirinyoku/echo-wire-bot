package bot

import (
	"context"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/kirinyoku/echo-wire-bot/internal/botkit"
	"github.com/kirinyoku/echo-wire-bot/internal/models"
)

// SourceStorage is an interface for adding a source to persistent storage.
// It provides the Add method, which saves a source and returns its ID or an error.
type SourceStorage interface {
	Add(ctx context.Context, source models.Source) (int64, error)
}

// ViewCmdAddSource creates a bot command handler for adding a new source.
// It parses the command arguments, adds the source to storage, and sends a confirmation message.
func ViewCmdAddSource(storage SourceStorage) botkit.ViewFunc {
	type addSourceArgs struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	}

	return func(ctx context.Context, bot *tgbotapi.BotAPI, update tgbotapi.Update) error {
		args, err := botkit.ParseJSON[addSourceArgs](update.Message.CommandArguments())
		if err != nil {
			return err
		}

		source := models.Source{
			Name: args.Name,
			URL:  args.URL,
		}

		sourceID, err := storage.Add(ctx, source)
		if err != nil {
			// TODO: send error message
			return err
		}

		var (
			msgText = fmt.Sprintf(
				"Source added with ID: `%d`\\. Use this ID to update the source or delete it.",
				sourceID,
			)
			reply = tgbotapi.NewMessage(update.Message.Chat.ID, msgText)
		)

		reply.ParseMode = parseModeMarkdownV2

		if _, err := bot.Send(reply); err != nil {
			return err
		}

		return nil
	}
}
