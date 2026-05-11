package providers

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Cache struct {
	Path string
}

func NewCache(path string) *Cache {
	return &Cache{Path: path}
}

func (c *Cache) Wrap(provider Provider) Provider {
	if c == nil || strings.TrimSpace(c.Path) == "" {
		return provider
	}
	return &cachedProvider{
		provider: provider,
		cache:    c,
	}
}

type cachedProvider struct {
	provider Provider
	cache    *Cache
}

func (p *cachedProvider) Name() string {
	return p.provider.Name()
}

func (p *cachedProvider) Search(ctx context.Context, query SearchQuery) ([]Candidate, error) {
	key, err := cacheKey(p.Name(), "search", query)
	if err != nil {
		return nil, err
	}

	var candidates []Candidate
	hit, err := p.cache.load(key, &candidates)
	if err != nil {
		return nil, err
	}
	if hit {
		return candidates, nil
	}

	candidates, err = p.provider.Search(ctx, query)
	if err != nil {
		return nil, err
	}
	if err := p.cache.store(key, candidates); err != nil {
		return nil, err
	}
	return candidates, nil
}

func (p *cachedProvider) Get(ctx context.Context, id string) (Candidate, error) {
	key, err := cacheKey(p.Name(), "get", strings.TrimSpace(id))
	if err != nil {
		return Candidate{}, err
	}

	var candidate Candidate
	hit, err := p.cache.load(key, &candidate)
	if err != nil {
		return Candidate{}, err
	}
	if hit {
		return candidate, nil
	}

	candidate, err = p.provider.Get(ctx, id)
	if err != nil {
		return Candidate{}, err
	}
	if err := p.cache.store(key, candidate); err != nil {
		return Candidate{}, err
	}
	return candidate, nil
}

func (c *Cache) load(key string, out any) (bool, error) {
	data, err := os.ReadFile(c.path(key))
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if err := json.Unmarshal(data, out); err != nil {
		return false, fmt.Errorf("read provider cache: %w", err)
	}
	return true, nil
}

func (c *Cache) store(key string, value any) error {
	path := c.path(key)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("write provider cache: %w", err)
	}
	return os.WriteFile(path, data, 0o644)
}

func (c *Cache) path(key string) string {
	return filepath.Join(c.Path, "providers", key+".json")
}

func cacheKey(providerName, operation string, value any) (string, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return "", fmt.Errorf("provider cache key: %w", err)
	}

	sum := sha256.Sum256(data)
	return strings.ToLower(providerName) + "-" + operation + "-" + hex.EncodeToString(sum[:]), nil
}
