package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/kirinyoku/echo-wire-bot/internal/models"
)

// ArticlePostgresStorage provides methods to interact with the articles table in a PostgreSQL database.
type ArticlePostgresStorage struct {
	db *sqlx.DB
}

// NewArticleStorage initializes a new instance of ArticlePostgresStorage.
func NewArticleStorage(db *sqlx.DB) *ArticlePostgresStorage {
	return &ArticlePostgresStorage{db: db}
}

// Store inserts a new article into the articles table. If the article already exists, it does nothing.
func (s *ArticlePostgresStorage) Store(ctx context.Context, article models.Article) error {
	const op = "storage.ArticlePostgresStorage.Store"

	conn, err := s.db.Connx(ctx)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	defer conn.Close()

	if _, err := conn.ExecContext(
		ctx,
		`INSERT INTO articles (source_id, title, link, summary, published_at)
	    				VALUES ($1, $2, $3, $4, $5)
	    				ON CONFLICT DO NOTHING;`,
		article.SourceID,
		article.Title,
		article.Link,
		article.Summary,
		article.PublishedAt,
	); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

// AllNotPosted retrieves articles that have not been marked as posted, filtered by a timestamp and limited by a maximum number.
func (s *ArticlePostgresStorage) AllNotPosted(ctx context.Context, since time.Time, limit uint64) ([]models.Article, error) {
	const op = "storage.ArticlePostgresStorage.AllNotPosted"

	conn, err := s.db.Connx(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	defer conn.Close()

	var dbArticles []dbArticleWithPriority

	if err := conn.SelectContext(
		ctx,
		&dbArticles,
		`SELECT 
				a.id AS a_id, 
				s.priority AS s_priority,
				s.id AS s_id,
				a.title AS a_title,
				a.link AS a_link,
				a.summary AS a_summary,
				a.published_at AS a_published_at,
				a.posted_at AS a_posted_at,
				a.created_at AS a_created_at
			FROM articles a JOIN sources s ON s.id = a.source_id
			WHERE a.posted_at IS NULL 
				AND a.published_at >= $1::timestamp
			ORDER BY a.created_at DESC, s_priority DESC LIMIT $2;`,
		since.UTC().Format(time.RFC3339),
		limit,
	); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	articles := (make([]models.Article, 0, len(dbArticles)))

	for _, dbArticle := range dbArticles {
		articles = append(articles, models.Article{
			ID:          dbArticle.ID,
			SourceID:    dbArticle.SourceID,
			Title:       dbArticle.Title,
			Link:        dbArticle.Link,
			Summary:     dbArticle.Summary.String,
			PublishedAt: dbArticle.PublishedAt,
			CreatedAt:   dbArticle.CreatedAt,
		})
	}

	return articles, nil
}

// MarkAsPosted updates the posted_at timestamp of an article, marking it as posted.
func (s *ArticlePostgresStorage) MarkAsPosted(ctx context.Context, article models.Article) error {
	const op = "storage.ArticlePostgresStorage.MarkAsPosted"

	conn, err := s.db.Connx(ctx)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	defer conn.Close()

	if _, err := conn.ExecContext(
		ctx,
		`UPDATE articles SET posted_at = $1::timestamp WHERE id = $2;`,
		time.Now().UTC().Format(time.RFC3339),
		article.ID,
	); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

// dbArticleWithPriority represents the structure of the database rows retrieved with additional source priority.
type dbArticleWithPriority struct {
	ID             int64          `db:"a_id"`
	SourcePriority int64          `db:"s_priority"`
	SourceID       int64          `db:"s_id"`
	Title          string         `db:"a_title"`
	Link           string         `db:"a_link"`
	Summary        sql.NullString `db:"a_summary"`
	PublishedAt    time.Time      `db:"a_published_at"`
	PostedAt       sql.NullTime   `db:"a_posted_at"`
	CreatedAt      time.Time      `db:"a_created_at"`
}
