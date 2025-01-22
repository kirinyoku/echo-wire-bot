package bot

import (
	"context"
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/kirinyoku/echo-wire-bot/internal/botkit"
	"github.com/kirinyoku/echo-wire-bot/internal/models"
)

// SourceLister is an interface for retrieving a list of sources from persistent storage.
// It defines the Sources method, which returns all sources.
type SourceLister interface {
	Sources(ctx context.Context) ([]models.Source, error)
}

// ViewCmdListSources creates a bot command handler for listing all sources.
// It retrieves the list of sources, formats their details, and sends the list as a message to the user.
func ViewCmdListSources(lister SourceLister) botkit.ViewFunc {
	return func(ctx context.Context, bot *tgbotapi.BotAPI, update tgbotapi.Update) error {
		sources, err := lister.Sources(ctx)
		if err != nil {
			return err
		}

		sourceInfos := make([]string, 0, len(sources))

		for _, source := range sources {
			sourceInfos = append(sourceInfos, formatSource(source))
		}

		msgText := fmt.Sprintf(
			"Source list\\(total %d\\):\n\n%s",
			len(sources),
			strings.Join(sourceInfos, "\n\n"),
		)

		reply := tgbotapi.NewMessage(update.Message.Chat.ID, msgText)
		reply.ParseMode = parseModeMarkdownV2

		if _, err := bot.Send(reply); err != nil {
			return err
		}

		return nil
	}
}
