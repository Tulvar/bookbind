package providers

import (
	"context"
	"fmt"
	"testing"
)

func TestRegistrySearchScoresAndSortsCandidates(t *testing.T) {
	registry := NewRegistry(stubProvider{candidates: []Candidate{
		{ID: "2", Title: "Дневной дозор", Authors: []string{"Сергей Лукьяненко"}},
		{ID: "1", Title: "Ночной дозор", Authors: []string{"Сергей Лукьяненко"}},
	}})

	got, err := registry.Search(context.Background(), SearchQuery{
		Title:  "Ночной дозор",
		Author: "Лукьяненко Сергей",
	})
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}

	if len(got) != 2 {
		t.Fatalf("len = %d, want 2", len(got))
	}
	if got[0].ID != "1" {
		t.Fatalf("first ID = %q, want 1", got[0].ID)
	}
	if got[0].Provider != "stub" {
		t.Fatalf("Provider = %q, want stub", got[0].Provider)
	}
	if got[0].Confidence <= got[1].Confidence {
		t.Fatalf("candidates were not sorted by confidence: %#v", got)
	}
}

func TestRegistryRejectsEmptyQuery(t *testing.T) {
	_, err := NewRegistry(stubProvider{}).Search(context.Background(), SearchQuery{})
	if err == nil {
		t.Fatal("Search() error = nil, want error")
	}
}

func TestRegistrySearchKeepsResultsWhenProviderFails(t *testing.T) {
	registry := NewRegistry(
		stubProvider{name: "broken", err: fmt.Errorf("too many requests")},
		stubProvider{name: "working", candidates: []Candidate{{ID: "book-1", Title: "Book"}}},
	)

	got, err := registry.Search(context.Background(), SearchQuery{Title: "Book"})
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("len = %d, want 1", len(got))
	}
	if got[0].Provider != "working" {
		t.Fatalf("Provider = %q, want working", got[0].Provider)
	}
}

func TestRegistrySearchReturnsErrorWhenAllProvidersFail(t *testing.T) {
	_, err := NewRegistry(stubProvider{name: "broken", err: fmt.Errorf("too many requests")}).
		Search(context.Background(), SearchQuery{Title: "Book"})
	if err == nil {
		t.Fatal("Search() error = nil, want error")
	}
}

func TestRegistryGetUsesSelectedProvider(t *testing.T) {
	got, err := NewRegistry(stubProvider{candidates: []Candidate{
		{ID: "book-1", Title: "Book"},
	}}).Get(context.Background(), "stub", "book-1")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if got.Title != "Book" {
		t.Fatalf("Title = %q", got.Title)
	}
	if got.Provider != "stub" {
		t.Fatalf("Provider = %q", got.Provider)
	}
}

func TestRegistryGetRejectsUnknownProvider(t *testing.T) {
	_, err := NewRegistry(stubProvider{}).Get(context.Background(), "unknown", "book-1")
	if err == nil {
		t.Fatal("Get() error = nil, want error")
	}
}

type stubProvider struct {
	name       string
	candidates []Candidate
	err        error
}

func (p stubProvider) Name() string {
	if p.name != "" {
		return p.name
	}
	return "stub"
}

func (p stubProvider) Search(context.Context, SearchQuery) ([]Candidate, error) {
	if p.err != nil {
		return nil, p.err
	}
	return p.candidates, nil
}

func (p stubProvider) Get(_ context.Context, id string) (Candidate, error) {
	for _, candidate := range p.candidates {
		if candidate.ID == id {
			return candidate, nil
		}
	}
	return Candidate{}, nil
}
