package local

import (
	"context"
	"testing"

	"github.com/Tulvar/bookbind/internal/providers"
)

func TestGet(t *testing.T) {
	provider := New([]providers.Candidate{{ID: "book-1", Title: "Book"}})

	got, err := provider.Get(context.Background(), "book-1")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if got.Title != "Book" {
		t.Fatalf("Title = %q", got.Title)
	}
}

func TestGetMissingCandidate(t *testing.T) {
	_, err := New(nil).Get(context.Background(), "missing")
	if err == nil {
		t.Fatal("Get() error = nil, want error")
	}
}
