package store

import (
	"context"
	"crypto/sha256"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Article struct {
	ID          int64
	URLHash     string
	URL         string
	Title       string
	Summary     string
	Author      string
	SourceID    int
	SourceName  string
	PublishedAt time.Time
	FetchedAt   time.Time
	Score       float64
	Category    string
}

type Source struct {
	ID                int
	Name              string
	Type              string
	URL               string
	Config            map[string]string
	Active            bool
	FetchIntervalMins int
	LastFetchedAt     *time.Time
}

type FeedQuery struct {
	Limit    int
	Offset   int
	SourceID int
	Search   string
}

type Store struct {
	pool *pgxpool.Pool
}

func New(ctx context.Context, dsn string) (*Store, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("pgxpool.New: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("ping: %w", err)
	}
	return &Store{pool: pool}, nil
}

func (s *Store) Close() {
	s.pool.Close()
}

func HashURL(rawURL string) string {
	normalized := strings.ToLower(strings.TrimRight(rawURL, "/"))
	h := sha256.Sum256([]byte(normalized))
	return fmt.Sprintf("%x", h)
}

func (s *Store) UpsertArticle(ctx context.Context, a Article) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO articles
			(url_hash, url, title, summary, author, source_id, published_at, fetched_at, score, category)
		VALUES
			($1, $2, $3, $4, $5, $6, $7, NOW(), $8, $9)
		ON CONFLICT (url_hash) DO NOTHING
	`, a.URLHash, a.URL, a.Title, a.Summary, a.Author, a.SourceID,
		a.PublishedAt, a.Score, a.Category)
	return err
}

func (s *Store) Feed(ctx context.Context, q FeedQuery) ([]Article, error) {
	if q.Limit == 0 {
		q.Limit = 20
	}

	args := pgx.NamedArgs{}
	where := "WHERE a.published_at > NOW() - INTERVAL '7 days'"

	if q.SourceID > 0 {
		where += " AND a.source_id = @source_id"
		args["source_id"] = q.SourceID
	}
	if q.Search != "" {
		where += " AND a.search_vec @@ plainto_tsquery('english', @search)"
		args["search"] = q.Search
	}

	args["limit"] = q.Limit
	args["offset"] = q.Offset

	rows, err := s.pool.Query(ctx, fmt.Sprintf(`
		SELECT a.id, a.url_hash, a.url, a.title, a.summary, a.author,
		       a.source_id, s.name, a.published_at, a.fetched_at, a.score, a.category
		FROM articles a
		JOIN sources s ON s.id = a.source_id
		%s
		ORDER BY a.score DESC, a.published_at DESC
		LIMIT @limit OFFSET @offset
	`, where), args)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}
	defer rows.Close()

	return scanArticles(rows)
}

func (s *Store) DailyDigest(ctx context.Context, limit int) ([]Article, error) {
	if limit == 0 {
		limit = 20
	}
	rows, err := s.pool.Query(ctx, `
		SELECT a.id, a.url_hash, a.url, a.title, a.summary, a.author,
		       a.source_id, s.name, a.published_at, a.fetched_at, a.score, a.category
		FROM articles a
		JOIN sources s ON s.id = a.source_id
		WHERE a.published_at > NOW() - INTERVAL '24 hours'
		ORDER BY a.score DESC, a.published_at DESC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}
	defer rows.Close()
	return scanArticles(rows)
}

func (s *Store) ActiveSources(ctx context.Context) ([]Source, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, name, type, url, config, active, fetch_interval_mins, last_fetched_at
		FROM sources
		WHERE active = true
	`)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}
	defer rows.Close()

	var sources []Source
	for rows.Next() {
		var src Source
		var cfg map[string]string
		if err := rows.Scan(
			&src.ID, &src.Name, &src.Type, &src.URL,
			&cfg, &src.Active, &src.FetchIntervalMins, &src.LastFetchedAt,
		); err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}
		src.Config = cfg
		sources = append(sources, src)
	}
	return sources, rows.Err()
}

func (s *Store) AllSources(ctx context.Context) ([]Source, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, name, type, url, config, active, fetch_interval_mins, last_fetched_at
		FROM sources ORDER BY name
	`)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}
	defer rows.Close()

	var sources []Source
	for rows.Next() {
		var src Source
		var cfg map[string]string
		if err := rows.Scan(
			&src.ID, &src.Name, &src.Type, &src.URL,
			&cfg, &src.Active, &src.FetchIntervalMins, &src.LastFetchedAt,
		); err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}
		src.Config = cfg
		sources = append(sources, src)
	}
	return sources, rows.Err()
}

func (s *Store) UpdateSourceFetchedAt(ctx context.Context, id int) error {
	_, err := s.pool.Exec(ctx,
		`UPDATE sources SET last_fetched_at = NOW() WHERE id = $1`, id)
	return err
}

func (s *Store) RunMigrations(ctx context.Context, migrations []string) error {
	for _, sql := range migrations {
		if _, err := s.pool.Exec(ctx, sql); err != nil {
			return fmt.Errorf("migration: %w", err)
		}
	}
	return nil
}

func scanArticles(rows pgx.Rows) ([]Article, error) {
	var articles []Article
	for rows.Next() {
		var a Article
		if err := rows.Scan(
			&a.ID, &a.URLHash, &a.URL, &a.Title, &a.Summary, &a.Author,
			&a.SourceID, &a.SourceName, &a.PublishedAt, &a.FetchedAt, &a.Score, &a.Category,
		); err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}
		articles = append(articles, a)
	}
	return articles, rows.Err()
}
