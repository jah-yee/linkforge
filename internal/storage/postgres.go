package storage

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/sekigo/linkforge/internal/domain"
)

var ErrNotFound = errors.New("not found")

type Postgres struct {
	pool *pgxpool.Pool
}

func NewPostgres(ctx context.Context, dsn string) (*Postgres, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse dsn: %w", err)
	}
	cfg.MaxConns = 10

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	return &Postgres{pool: pool}, nil
}

func (p *Postgres) Close() {
	p.pool.Close()
}

func (p *Postgres) NextSequence(ctx context.Context) (int64, error) {
	var n int64
	err := p.pool.QueryRow(ctx, `SELECT nextval('link_seq')`).Scan(&n)
	if err != nil {
		return 0, fmt.Errorf("nextval: %w", err)
	}
	return n, nil
}

func (p *Postgres) InsertLink(ctx context.Context, code, rawURL string) (*domain.Link, error) {
	var link domain.Link
	err := p.pool.QueryRow(ctx, `
		INSERT INTO links (code, url)
		VALUES ($1, $2)
		RETURNING id, code, url, created_at
	`, code, rawURL).Scan(&link.ID, &link.Code, &link.URL, &link.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("insert link: %w", err)
	}
	return &link, nil
}

func (p *Postgres) GetLinkByCode(ctx context.Context, code string) (*domain.Link, error) {
	var link domain.Link
	err := p.pool.QueryRow(ctx, `
		SELECT id, code, url, created_at FROM links WHERE code = $1
	`, code).Scan(&link.ID, &link.Code, &link.URL, &link.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("query link: %w", err)
	}
	return &link, nil
}
