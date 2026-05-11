package openlibrary

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Tulvar/bookbind/internal/providers"
)

func TestSearch(t *testing.T) {
	var gotPath string
	var gotQuery string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"docs": [
				{
					"key": "/works/OL1W",
					"title": "Ночной дозор",
					"author_name": ["Сергей Лукьяненко"],
					"first_publish_year": 1998,
					"cover_i": 12345
				}
			]
		}`))
	}))
	defer server.Close()

	provider := New(
		WithBaseURL(server.URL),
		WithHTTPClient(server.Client()),
		WithLimit(5),
	)

	got, err := provider.Search(context.Background(), providers.SearchQuery{
		Title:  "Ночной дозор",
		Author: "Лукьяненко",
	})
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}

	if gotPath != "/search.json" {
		t.Fatalf("path = %q, want /search.json", gotPath)
	}
	for _, want := range []string{
		"title=%D0%9D%D0%BE%D1%87%D0%BD%D0%BE%D0%B9",
		"author=%D0%9B%D1%83%D0%BA%D1%8C%D1%8F%D0%BD%D0%B5%D0%BD%D0%BA%D0%BE",
		"limit=5",
		"fields=",
	} {
		if !strings.Contains(gotQuery, want) {
			t.Fatalf("query %q does not contain %q", gotQuery, want)
		}
	}

	if len(got) != 1 {
		t.Fatalf("len = %d, want 1", len(got))
	}
	candidate := got[0]
	if candidate.Provider != "openlibrary" {
		t.Fatalf("Provider = %q", candidate.Provider)
	}
	if candidate.ID != "/works/OL1W" {
		t.Fatalf("ID = %q", candidate.ID)
	}
	if candidate.Title != "Ночной дозор" {
		t.Fatalf("Title = %q", candidate.Title)
	}
	if candidate.Authors[0] != "Сергей Лукьяненко" {
		t.Fatalf("Author = %q", candidate.Authors[0])
	}
	if candidate.Year != 1998 {
		t.Fatalf("Year = %d", candidate.Year)
	}
	if candidate.CoverURL != "https://covers.openlibrary.org/b/id/12345-L.jpg" {
		t.Fatalf("CoverURL = %q", candidate.CoverURL)
	}
}

func TestSearchRejectsEmptyQuery(t *testing.T) {
	_, err := New().Search(context.Background(), providers.SearchQuery{})
	if err == nil {
		t.Fatal("Search() error = nil, want error")
	}
}

func TestSearchHandlesHTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "nope", http.StatusBadGateway)
	}))
	defer server.Close()

	_, err := New(WithBaseURL(server.URL), WithHTTPClient(server.Client())).Search(
		context.Background(),
		providers.SearchQuery{Title: "Book"},
	)
	if err == nil {
		t.Fatal("Search() error = nil, want error")
	}
}

func TestCoverURL(t *testing.T) {
	if coverURL(0) != "" {
		t.Fatal("coverURL(0) should be empty")
	}
	if got, want := coverURL(42), "https://covers.openlibrary.org/b/id/42-L.jpg"; got != want {
		t.Fatalf("coverURL() = %q, want %q", got, want)
	}
}
