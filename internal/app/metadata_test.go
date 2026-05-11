package app

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Tulvar/bookbind/internal/metadata"
	"github.com/Tulvar/bookbind/internal/providers"
	"github.com/Tulvar/bookbind/internal/providers/local"
)

func TestResolveMetadataWritesYAML(t *testing.T) {
	app := New(WithProviders(providers.NewRegistry(local.New([]providers.Candidate{
		{
			ID:          "book-1",
			Title:       "Ночной дозор",
			Authors:     []string{"Сергей Лукьяненко"},
			Series:      "Дозоры",
			SeriesIndex: "1",
			Year:        1998,
			CoverURL:    "https://example.test/cover.jpg",
		},
	}))))
	outputPath := filepath.Join(t.TempDir(), "bookbind.yaml")

	result, err := app.ResolveMetadata(context.Background(), ResolveMetadataRequest{
		Provider:   "local",
		ID:         "book-1",
		OutputPath: outputPath,
	})
	if err != nil {
		t.Fatalf("ResolveMetadata() error = %v", err)
	}

	if result.Book.Title != "Ночной дозор" {
		t.Fatalf("Title = %q", result.Book.Title)
	}
	book, err := metadata.LoadYAML(outputPath)
	if err != nil {
		t.Fatalf("LoadYAML() error = %v", err)
	}
	if book.Title != "Ночной дозор" {
		t.Fatalf("Title = %q", book.Title)
	}
	if book.Author != "Сергей Лукьяненко" {
		t.Fatalf("Author = %q", book.Author)
	}
	if book.Series != "Дозоры" || book.SeriesIndex != "1" {
		t.Fatalf("Series = %q, index = %q", book.Series, book.SeriesIndex)
	}
	if book.PublishedYear != 1998 {
		t.Fatalf("PublishedYear = %d", book.PublishedYear)
	}
}

func TestResolveMetadataRejectsExistingOutput(t *testing.T) {
	app := New(WithProviders(providers.NewRegistry(local.New([]providers.Candidate{
		{ID: "book-1", Title: "Book"},
	}))))
	outputPath := filepath.Join(t.TempDir(), "bookbind.yaml")
	if err := os.WriteFile(outputPath, []byte("existing"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	_, err := app.ResolveMetadata(context.Background(), ResolveMetadataRequest{
		Provider:   "local",
		ID:         "book-1",
		OutputPath: outputPath,
	})
	if err == nil {
		t.Fatal("ResolveMetadata() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "output already exists") {
		t.Fatalf("ResolveMetadata() error = %v", err)
	}
}

func TestResolveMetadataRejectsUnknownProvider(t *testing.T) {
	app := New()

	_, err := app.ResolveMetadata(context.Background(), ResolveMetadataRequest{
		Provider: "unknown",
		ID:       "book-1",
	})
	if err == nil {
		t.Fatal("ResolveMetadata() error = nil, want error")
	}
}

func TestBookFromCandidateJoinsMultipleAuthors(t *testing.T) {
	book := bookFromCandidate(providers.Candidate{
		Title:   "Book",
		Authors: []string{"One", "Two"},
	})

	if book.Author != "One, Two" {
		t.Fatalf("Author = %q", book.Author)
	}
	if len(book.Authors) != 2 {
		t.Fatalf("Authors len = %d", len(book.Authors))
	}
}
