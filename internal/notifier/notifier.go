package notifier

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/go-shiori/go-readability"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/kirinyoku/echo-wire-bot/internal/botkit/markup"
	"github.com/kirinyoku/echo-wire-bot/internal/models"
)

// ArticleProvider defines the interface for working with articles.
type ArticleProvider interface {
	// AllNotPosted retrieves articles that have not been posted yet,
	// filtered by a timestamp and limited by a specified number.
	AllNotPosted(ctx context.Context, since time.Time, limit uint64) ([]models.Article, error)
	// MarkAsPosted updates an article to indicate it has been posted.
	MarkAsPosted(ctx context.Context, article models.Article) error
}

// Summarizer defines the interface for generating summaries of text content.
type Summarizer interface {
	// Summarize generates a summary for the provided text.
	Summarize(text string) (string, error)
}

// Notifier handles the process of selecting, summarizing, and sending articles
// to a specified Telegram channel at regular intervals.
type Notifier struct {
	articles         ArticleProvider
	summarizer       Summarizer
	bot              *tgbotapi.BotAPI
	sendInterval     time.Duration
	lookupTimeWindow time.Duration
	channelID        int64
}

// New initializes and returns a new Notifier instance.
func New(
	articleProvider ArticleProvider,
	summarizer Summarizer,
	bot *tgbotapi.BotAPI,
	sendInterval time.Duration,
	lookupTimeWindow time.Duration,
	channelID int64,
) *Notifier {
	return &Notifier{
		articles:         articleProvider,
		summarizer:       summarizer,
		bot:              bot,
		sendInterval:     sendInterval,
		lookupTimeWindow: lookupTimeWindow,
		channelID:        channelID,
	}
}

// Start begins the Notifier's routine to periodically send articles.
// It stops when the context is canceled or an error occurs.
func (n *Notifier) Start(ctx context.Context) error {
	ticker := time.NewTicker(n.sendInterval)
	defer ticker.Stop()

	if err := n.SelectAndSendArticle(ctx); err != nil {
		return err
	}

	for {
		select {
		case <-ticker.C:
			if err := n.SelectAndSendArticle(ctx); err != nil {
				return err
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// SelectAndSendArticle selects the top article, generates a summary if needed,
// sends the article to the Telegram channel, and marks it as posted.
func (n *Notifier) SelectAndSendArticle(ctx context.Context) error {
	topOneArticles, err := n.articles.AllNotPosted(ctx, time.Now().Add(-n.lookupTimeWindow), 1)
	if err != nil {
		return err
	}

	if len(topOneArticles) == 0 {
		return nil
	}

	article := topOneArticles[0]

	summary, err := n.extractSummary(article)
	if err != nil {
		log.Printf("[ERROR] failed to extract summary: %v", err)
	}

	if err := n.sendArticle(article, summary); err != nil {
		return err
	}

	return n.articles.MarkAsPosted(ctx, article)
}

var redundantNewLines = regexp.MustCompile(`\n{3,}`)

// extractSummary retrieves or generates a summary for the given article.
func (n *Notifier) extractSummary(article models.Article) (string, error) {
	var r io.Reader

	if article.Summary != "" {
		r = strings.NewReader(article.Summary)
	} else {
		resp, err := http.Get(article.Link)
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()

		r = resp.Body
	}

	doc, err := readability.FromReader(r, nil)
	if err != nil {
		return "", err
	}

	summary, err := n.summarizer.Summarize(cleanupText(doc.TextContent))
	if err != nil {
		return "", err
	}

	return "\n\n" + summary, nil
}

// cleanupText removes redundant newlines from the provided text.
func cleanupText(text string) string {
	return redundantNewLines.ReplaceAllString(text, "\n")
}

// sendArticle sends the article with its summary to the Telegram channel.
func (n *Notifier) sendArticle(article models.Article, summary string) error {
	const msgFormat = "*%s*%s\n\n%s"

	msg := tgbotapi.NewMessage(n.channelID, fmt.Sprintf(
		msgFormat,
		markup.EscapeForMarkdown(article.Title),
		markup.EscapeForMarkdown(summary),
		markup.EscapeForMarkdown(article.Link),
	))
	msg.ParseMode = "MarkdownV2"

	_, err := n.bot.Send(msg)
	if err != nil {
		return err
	}

	return nil
}
