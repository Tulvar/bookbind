package metadata

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadYAML(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bookbind.yaml")
	data := []byte(`title: "Night Watch"
author: "Sergey Lukyanenko"
narrator: "Reader"
series: "Watches"
series_index: "1"
language: "ru"
genre: "Fantasy"
publisher: "Publisher"
published_year: 1998
description: |
  Line one.
  Line two.
`)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("write metadata: %v", err)
	}

	book, err := LoadYAML(path)
	if err != nil {
		t.Fatalf("LoadYAML() error = %v", err)
	}

	if book.Title != "Night Watch" {
		t.Fatalf("Title = %q", book.Title)
	}
	if got := book.NormalizedAuthors(); len(got) != 1 || got[0] != "Sergey Lukyanenko" {
		t.Fatalf("authors = %#v", got)
	}
	if book.Description == "" {
		t.Fatal("Description is empty")
	}
}

func TestBookEmpty(t *testing.T) {
	empty := Book{}
	withTitle := Book{Title: "Book"}

	if !empty.Empty() {
		t.Fatal("empty book was not empty")
	}
	if withTitle.Empty() {
		t.Fatal("book with title was empty")
	}
}
