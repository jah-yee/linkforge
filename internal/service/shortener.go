package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"strings"

	"github.com/sekigo/linkforge/internal/domain"
	"github.com/sekigo/linkforge/internal/storage"
)

const base62Alphabet = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

var (
	ErrInvalidURL = errors.New("invalid url")
	ErrNotFound   = errors.New("not found")
)

type Storage interface {
	NextSequence(ctx context.Context) (int64, error)
	InsertLink(ctx context.Context, code, rawURL string) (*domain.Link, error)
	GetLinkByCode(ctx context.Context, code string) (*domain.Link, error)
}

type CacheStore interface {
	GetURL(ctx context.Context, code string) (string, error)
	SetURL(ctx context.Context, code, rawURL string) error
}

type Shortener struct {
	db      Storage
	cache   CacheStore
	baseURL string
}

func NewShortener(db Storage, cache CacheStore, baseURL string) *Shortener {
	return &Shortener{db: db, cache: cache, baseURL: baseURL}
}

func (s *Shortener) BaseURL() string {
	return s.baseURL
}

func (s *Shortener) Shorten(ctx context.Context, rawURL string) (*domain.Link, error) {
	if err := validateURL(rawURL); err != nil {
		return nil, err
	}

	seq, err := s.db.NextSequence(ctx)
	if err != nil {
		return nil, fmt.Errorf("next seq: %w", err)
	}
	code := encodeBase62(seq)

	link, err := s.db.InsertLink(ctx, code, rawURL)
	if err != nil {
		return nil, err
	}

	if err := s.cache.SetURL(ctx, code, rawURL); err != nil {
		slog.WarnContext(ctx, "cache set failed", "code", code, "err", err)
	}

	return link, nil
}

func (s *Shortener) Resolve(ctx context.Context, code string) (string, error) {
	if v, err := s.cache.GetURL(ctx, code); err == nil {
		return v, nil
	} else if !errors.Is(err, storage.ErrNotFound) {
		slog.WarnContext(ctx, "cache get failed", "code", code, "err", err)
	}

	link, err := s.db.GetLinkByCode(ctx, code)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return "", ErrNotFound
		}
		return "", err
	}

	if err := s.cache.SetURL(ctx, code, link.URL); err != nil {
		slog.WarnContext(ctx, "cache backfill failed", "code", code, "err", err)
	}

	return link.URL, nil
}

func validateURL(raw string) error {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return fmt.Errorf("%w: empty", ErrInvalidURL)
	}
	u, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidURL, err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("%w: scheme must be http or https", ErrInvalidURL)
	}
	if u.Host == "" {
		return fmt.Errorf("%w: host required", ErrInvalidURL)
	}
	return nil
}

func encodeBase62(n int64) string {
	if n == 0 {
		return string(base62Alphabet[0])
	}
	var b []byte
	for n > 0 {
		b = append([]byte{base62Alphabet[n%62]}, b...)
		n /= 62
	}
	return string(b)
}
