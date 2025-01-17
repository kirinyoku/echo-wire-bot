package fetcher

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/kirinyoku/echo-wire-bot/internal/models"
	"github.com/kirinyoku/echo-wire-bot/internal/source"
)

// ArticleStorage defines the interface for storing articles in a persistent layer.
type ArticleStorage interface {
	Store(context.Context, models.Article) error
}

// SourceProvider defines the interface for fetching a list of sources from a storage layer.
type SourceProvider interface {
	Sources(ctx context.Context) ([]models.Source, error)
}

// Source defines the interface for an individual source, including fetching items and metadata.
type Source interface {
	ID() int64
	Name() string
	Fetch(ctx context.Context) ([]models.Item, error)
}

// Fetcher is responsible for fetching and processing articles from multiple sources at regular intervals.
type Fetcher struct {
	articles ArticleStorage
	source   SourceProvider

	fetchInterval  time.Duration
	filterKeywords []string
}

// New creates a new Fetcher instance with the provided dependencies and configuration.
func New(
	articles ArticleStorage,
	source SourceProvider,
	fetchIntrerval time.Duration,
	filterKeywords []string,
) *Fetcher {
	return &Fetcher{
		articles:       articles,
		source:         source,
		fetchInterval:  fetchIntrerval,
		filterKeywords: filterKeywords,
	}
}

// Run starts the Fetcher to periodically fetch articles from sources.
// It blocks until the context is canceled or an error occurs during the first fetch.
func (f *Fetcher) Run(ctx context.Context) error {
	const op = "fetcher.Run"

	ticker := time.NewTicker(f.fetchInterval)
	defer ticker.Stop()

	if err := f.Fetch(ctx); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			if err := f.Fetch(ctx); err != nil {
				return fmt.Errorf("%s: %w", op, err)
			}
		}
	}
}

// Fetch retrieves and processes items from all sources concurrently.
func (f *Fetcher) Fetch(ctx context.Context) error {
	const op = "fetcher.Fetch"

	sources, err := f.source.Sources(ctx)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	var wg sync.WaitGroup
	errCh := make(chan error, len(sources))

	for _, s := range sources {
		wg.Add(1)

		rssSource := source.NewRSSSource(s)

		go func(src Source) {
			defer wg.Done()

			items, err := src.Fetch(ctx)
			if err != nil {
				errCh <- fmt.Errorf("fetching source %s failed: %w", src.Name(), err)
				return
			}

			if err := f.processItem(ctx, src, items); err != nil {
				errCh <- fmt.Errorf("processing items for source %s failed: %w", src.Name(), err)
				return
			}
		}(rssSource)
	}

	wg.Wait()
	close(errCh)

	for fetchErr := range errCh {
		log.Printf("error: %v", fetchErr)
	}

	return nil
}

// processItem processes a batch of items fetched from a single source.
// It stores valid items in the storage layer.
func (f *Fetcher) processItem(ctx context.Context, s Source, items []models.Item) error {
	const op = "fetcher.processItem"

	for _, item := range items {
		item.Date = item.Date.UTC()

		if f.itemShouldBeSkipped(item) {
			continue
		}

		if err := f.articles.Store(ctx, models.Article{
			SourceID:    s.ID(),
			Title:       item.Title,
			Link:        item.Link,
			Summary:     item.Summary,
			PublishedAt: item.Date,
		}); err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
	}

	return nil
}

// itemShouldBeSkipped determines whether an item should be filtered out based on its title or categories.
func (f *Fetcher) itemShouldBeSkipped(item models.Item) bool {
	title := strings.ToLower(item.Title)

	for _, keyword := range f.filterKeywords {
		keyword = strings.ToLower(keyword)
		if strings.Contains(title, keyword) {
			return true
		}

		for _, category := range item.Categories {
			if strings.EqualFold(category, keyword) {
				return true
			}
		}
	}

	return false
}
