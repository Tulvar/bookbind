package providers

import (
	"context"
	"testing"
)

func TestCachedProviderCachesSearch(t *testing.T) {
	provider := &countingProvider{
		searchCandidates: []Candidate{{ID: "book-1", Title: "Book"}},
	}
	cached := NewCache(t.TempDir()).Wrap(provider)

	query := SearchQuery{Title: "Book"}
	first, err := cached.Search(context.Background(), query)
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	second, err := cached.Search(context.Background(), query)
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}

	if provider.searchCalls != 1 {
		t.Fatalf("searchCalls = %d, want 1", provider.searchCalls)
	}
	if len(first) != 1 || len(second) != 1 {
		t.Fatalf("results len = %d/%d", len(first), len(second))
	}
}

func TestCachedProviderCachesGet(t *testing.T) {
	provider := &countingProvider{
		getCandidate: Candidate{ID: "book-1", Title: "Book"},
	}
	cached := NewCache(t.TempDir()).Wrap(provider)

	first, err := cached.Get(context.Background(), "book-1")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	second, err := cached.Get(context.Background(), "book-1")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if provider.getCalls != 1 {
		t.Fatalf("getCalls = %d, want 1", provider.getCalls)
	}
	if first.ID != "book-1" || second.ID != "book-1" {
		t.Fatalf("IDs = %q/%q", first.ID, second.ID)
	}
}

func TestCacheWrapAllowsEmptyPath(t *testing.T) {
	provider := &countingProvider{}

	wrapped := NewCache("").Wrap(provider)

	if wrapped != provider {
		t.Fatal("Wrap() should return original provider for empty path")
	}
}

type countingProvider struct {
	searchCandidates []Candidate
	getCandidate     Candidate
	searchCalls      int
	getCalls         int
}

func (p *countingProvider) Name() string {
	return "counting"
}

func (p *countingProvider) Search(context.Context, SearchQuery) ([]Candidate, error) {
	p.searchCalls++
	return p.searchCandidates, nil
}

func (p *countingProvider) Get(context.Context, string) (Candidate, error) {
	p.getCalls++
	return p.getCandidate, nil
}
