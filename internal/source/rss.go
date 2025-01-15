package source

import (
	"context"
	"fmt"

	"github.com/SlyMarbo/rss"
	"github.com/kirinyoku/echo-wire-bot/internal/models"
)

// RSSSource represents a specific RSS feed source.
type RSSSource struct {
	URL        string
	SourceID   int64
	SourceName string
}

// NewRSSSourcel creates a new instance of RSSSource from a Source model.
func NewRSSSource(m models.Source) RSSSource {
	return RSSSource{
		URL:        m.URL,
		SourceID:   m.ID,
		SourceName: m.Name,
	}
}

// Fetch retrieves and parses items from the RSS feed.
// It uses a context to handle timeouts or cancellations.
func (s *RSSSource) Fetch(ctx context.Context) ([]models.Item, error) {
	const op = "source.RSSSource.Fetch"

	feed, err := s.loadFeed(ctx, s.URL)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	items := make([]models.Item, 0, len(feed.Items))

	for _, item := range feed.Items {
		items = append(items, models.Item{
			Title:      item.Title,
			Categories: item.Categories,
			Link:       item.Link,
			Date:       item.Date,
			Summary:    item.Summary,
			SourceName: s.SourceName,
		})
	}

	return items, nil
}

// loadFeed asynchronously fetches the RSS feed and returns it.
func (s *RSSSource) loadFeed(ctx context.Context, url string) (*rss.Feed, error) {
	var (
		feedCh = make(chan *rss.Feed)
		errCh  = make(chan error)
	)

	go func() {
		feed, err := rss.Fetch(url)
		if err != nil {
			errCh <- <-errCh
			return
		}

		feedCh <- feed
	}()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case err := <-errCh:
		return nil, err
	case feed := <-feedCh:
		return feed, nil
	}
}
