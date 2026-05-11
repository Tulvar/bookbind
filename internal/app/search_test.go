package app

import (
	"context"
	"testing"

	"github.com/Tulvar/bookbind/internal/providers"
	"github.com/Tulvar/bookbind/internal/providers/local"
)

func TestSearchMetadata(t *testing.T) {
	app := New(WithProviders(providers.NewRegistry(local.New([]providers.Candidate{
		{ID: "1", Title: "Ночной дозор", Authors: []string{"Сергей Лукьяненко"}},
	}))))

	got, err := app.SearchMetadata(context.Background(), SearchRequest{
		Title:  "Ночной дозор",
		Author: "Лукьяненко Сергей",
	})
	if err != nil {
		t.Fatalf("SearchMetadata() error = %v", err)
	}

	if len(got.Candidates) != 1 {
		t.Fatalf("len = %d, want 1", len(got.Candidates))
	}
	if got.Candidates[0].Confidence != 1 {
		t.Fatalf("Confidence = %v, want 1", got.Candidates[0].Confidence)
	}
}
