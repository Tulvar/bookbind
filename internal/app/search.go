package app

import (
	"context"
	"fmt"

	"github.com/Tulvar/bookbind/internal/providers"
)

type SearchRequest struct {
	Title     string
	Author    string
	Providers []string
}

type SearchResult struct {
	Candidates []providers.Candidate
}

func (a *App) SearchMetadata(ctx context.Context, req SearchRequest) (SearchResult, error) {
	if a.providers == nil {
		return SearchResult{}, fmt.Errorf("metadata providers are not configured")
	}

	registry := a.providers
	if len(req.Providers) > 0 {
		selected, err := NewProviderRegistry(req.Providers)
		if err != nil {
			return SearchResult{}, err
		}
		registry = selected
	}

	candidates, err := registry.Search(ctx, providers.SearchQuery{
		Title:  req.Title,
		Author: req.Author,
	})
	if err != nil {
		return SearchResult{}, err
	}
	return SearchResult{Candidates: candidates}, nil
}
