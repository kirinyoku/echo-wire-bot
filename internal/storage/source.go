package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/kirinyoku/echo-wire-bot/internal/models"
)

// SourcePostgresStorage provides storage for RSS sources using a PostgreSQL database.
type SourcePostgresStorage struct {
	db *sqlx.DB
}

// dbSource maps database rows to Go structs for internal use.
type dbSource struct {
	ID        int64     `db:"id"`
	Name      string    `db:"name"`
	URL       string    `db:"url"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

// Sources retrieves all sources from the database.
func (s *SourcePostgresStorage) Sources(ctx context.Context) ([]models.Source, error) {
	const op = "storage.SourcePostgresStorage.Sources"

	conn, err := s.db.Connx(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	defer conn.Close()

	var sourcesDB []dbSource
	var sources []models.Source

	if err := conn.SelectContext(ctx, &sourcesDB, "SELECT * FROM sources"); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	for _, sourceDB := range sourcesDB {
		sources = append(sources, models.Source{
			ID:        sourceDB.ID,
			Name:      sourceDB.Name,
			URL:       sourceDB.URL,
			CreatedAt: sourceDB.CreatedAt,
		})
	}

	return sources, nil
}

// SourceByID retrieves a source by its ID.
func (s *SourcePostgresStorage) SourceByID(ctx context.Context, id int64) (models.Source, error) {
	const op = "storage.SourcePostgresStorage.SourceByID"

	conn, err := s.db.Connx(ctx)
	if err != nil {
		return models.Source{}, fmt.Errorf("%s: %w", op, err)
	}

	defer conn.Close()

	var sourceDB dbSource

	if err := conn.GetContext(ctx, &sourceDB, "SELECT * FROM sources WHERE id = $1", id); err != nil {
		return models.Source{}, fmt.Errorf("%s: %w", op, err)
	}

	return models.Source{
		ID:        sourceDB.ID,
		Name:      sourceDB.Name,
		URL:       sourceDB.URL,
		CreatedAt: sourceDB.CreatedAt,
	}, nil
}

// Add inserts a new source into the database and returns its ID.
func (s *SourcePostgresStorage) Add(ctx context.Context, source models.Source) (int64, error) {
	const op = "storage.SourcePostgresStorage.Add"

	conn, err := s.db.Connx(ctx)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	defer conn.Close()

	var id int64

	row := conn.QueryRowContext(ctx, "INSERT INTO sources (name, url) VALUES ($1, $2)", source.Name, source.URL)

	if err := row.Err(); err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	if err = row.Scan(&id); err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

// Delete removes a source from the database by ID.
func (s *SourcePostgresStorage) Delete(ctx context.Context, id int64) error {
	const op = "storage.SourcePostgresStorage.Delete"

	conn, err := s.db.Connx(ctx)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	defer conn.Close()

	if _, err := conn.ExecContext(ctx, "DELETE FROM sources WHERE id = $1", id); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
