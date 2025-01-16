package models

import "time"

// Item represents an RSS feed item.
type Item struct {
	Title      string
	Categories []string
	Link       string
	Date       time.Time
	Summary    string
	SourceName string
}

// Source represents an RSS feed source.
type Source struct {
	ID        int64
	Name      string
	URL       string
	CreatedAt time.Time
}

// Article represents an individual article fetched from an RSS feed.
type Article struct {
	ID          int64
	SourceID    int64
	Title       string
	Link        string
	Summary     string
	PublishedAt time.Time
	PostedAt    time.Time
	CreatedAt   time.Time
}
