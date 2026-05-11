package providers

import (
	"context"
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

type stubProvider struct {
	candidates []Candidate
}

func (p stubProvider) Name() string {
	return "stub"
}

func (p stubProvider) Search(context.Context, SearchQuery) ([]Candidate, error) {
	return p.candidates, nil
}

func (p stubProvider) Get(context.Context, string) (Candidate, error) {
	return Candidate{}, nil
}
