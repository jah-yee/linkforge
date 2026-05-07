package service

import (
	"context"
	"errors"
	"testing"

	"github.com/sekigo/linkforge/internal/domain"
	"github.com/sekigo/linkforge/internal/storage"
)

func TestEncodeBase62(t *testing.T) {
	tests := []struct {
		in   int64
		want string
	}{
		{0, "0"},
		{1, "1"},
		{10, "a"},
		{61, "Z"},
		{62, "10"},
		{3843, "ZZ"},
	}
	for _, tt := range tests {
		if got := encodeBase62(tt.in); got != tt.want {
			t.Errorf("encodeBase62(%d) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestValidateURL(t *testing.T) {
	valid := []string{
		"http://example.com",
		"https://example.com/path?q=1",
		"https://sub.domain.example.com/a/b",
	}
	for _, u := range valid {
		if err := validateURL(u); err != nil {
			t.Errorf("validateURL(%q) unexpected err: %v", u, err)
		}
	}
	invalid := []string{
		"",
		"   ",
		"ftp://example.com",
		"https://",
		"javascript:alert(1)",
	}
	for _, u := range invalid {
		if err := validateURL(u); err == nil {
			t.Errorf("validateURL(%q) expected error", u)
		}
	}
}

type fakeStore struct {
	seq   int64
	links map[string]*domain.Link
}

func newFakeStore() *fakeStore {
	return &fakeStore{links: map[string]*domain.Link{}}
}

func (f *fakeStore) NextSequence(_ context.Context) (int64, error) {
	f.seq++
	return f.seq, nil
}

func (f *fakeStore) InsertLink(_ context.Context, code, rawURL string) (*domain.Link, error) {
	l := &domain.Link{ID: f.seq, Code: code, URL: rawURL}
	f.links[code] = l
	return l, nil
}

func (f *fakeStore) GetLinkByCode(_ context.Context, code string) (*domain.Link, error) {
	l, ok := f.links[code]
	if !ok {
		return nil, storage.ErrNotFound
	}
	return l, nil
}

type fakeCache struct{ data map[string]string }

func newFakeCache() *fakeCache { return &fakeCache{data: map[string]string{}} }

func (c *fakeCache) GetURL(_ context.Context, code string) (string, error) {
	v, ok := c.data[code]
	if !ok {
		return "", storage.ErrNotFound
	}
	return v, nil
}

func (c *fakeCache) SetURL(_ context.Context, code, url string) error {
	c.data[code] = url
	return nil
}

func TestShortenerRoundtrip(t *testing.T) {
	ctx := context.Background()
	sh := NewShortener(newFakeStore(), newFakeCache(), "http://x")

	link, err := sh.Shorten(ctx, "https://example.com")
	if err != nil {
		t.Fatalf("shorten: %v", err)
	}
	if link.Code == "" {
		t.Fatal("expected non-empty code")
	}

	got, err := sh.Resolve(ctx, link.Code)
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if got != "https://example.com" {
		t.Errorf("got %q, want https://example.com", got)
	}

	if _, err := sh.Resolve(ctx, "nope"); !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestShortenerRejectsBadURL(t *testing.T) {
	sh := NewShortener(newFakeStore(), newFakeCache(), "http://x")
	_, err := sh.Shorten(context.Background(), "not a url")
	if !errors.Is(err, ErrInvalidURL) {
		t.Errorf("expected ErrInvalidURL, got %v", err)
	}
}
