package bot

import (
	"context"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/kirinyoku/echo-wire-bot/internal/botkit"
)

// SourceDeleter is an interface for deleting a source from persistent storage.
// It provides the Delete method, which removes a source by its ID and returns an error if unsuccessful.
type SourceDeleter interface {
	Delete(ctx context.Context, sourceID int64) error
}

// ViewCmdDeleteSource creates a bot command handler for deleting a source.
// It parses the source ID from the command arguments, deletes the source from storage,
// and sends a confirmation message to the user.
func ViewCmdDeleteSource(deleter SourceDeleter) botkit.ViewFunc {
	return func(ctx context.Context, bot *tgbotapi.BotAPI, update tgbotapi.Update) error {
		idStr := update.Message.CommandArguments()

		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			return err
		}

		if err := deleter.Delete(ctx, id); err != nil {
			return nil
		}

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "The source has been successfully removed")
		if _, err := bot.Send(msg); err != nil {
			return err
		}

		return nil
	}
}
