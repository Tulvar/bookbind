package googlebooks

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
			"items": [
				{
					"id": "volume-1",
					"volumeInfo": {
						"title": "The Hobbit",
						"subtitle": "There and Back Again",
						"authors": ["J. R. R. Tolkien"],
						"publisher": "Allen & Unwin",
						"publishedDate": "1937-09-21",
						"description": "A hobbit goes on an adventure.",
						"categories": ["Fantasy", "Adventure"],
						"language": "en",
						"imageLinks": {
							"thumbnail": "http://example.test/thumb.jpg",
							"large": "https://example.test/large.jpg"
						}
					}
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
		Title:  "The Hobbit",
		Author: "Tolkien",
	})
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}

	if gotPath != "/volumes" {
		t.Fatalf("path = %q, want /volumes", gotPath)
	}
	for _, want := range []string{
		"q=",
		"intitle%3A%22The+Hobbit%22",
		"inauthor%3ATolkien",
		"maxResults=5",
		"projection=lite",
	} {
		if !strings.Contains(gotQuery, want) {
			t.Fatalf("query %q does not contain %q", gotQuery, want)
		}
	}

	if len(got) != 1 {
		t.Fatalf("len = %d, want 1", len(got))
	}
	candidate := got[0]
	if candidate.Provider != "googlebooks" {
		t.Fatalf("Provider = %q", candidate.Provider)
	}
	if candidate.ID != "volume-1" {
		t.Fatalf("ID = %q", candidate.ID)
	}
	if candidate.Title != "The Hobbit" {
		t.Fatalf("Title = %q", candidate.Title)
	}
	if candidate.Subtitle != "There and Back Again" {
		t.Fatalf("Subtitle = %q", candidate.Subtitle)
	}
	if candidate.Authors[0] != "J. R. R. Tolkien" {
		t.Fatalf("Author = %q", candidate.Authors[0])
	}
	if candidate.Year != 1937 {
		t.Fatalf("Year = %d", candidate.Year)
	}
	if candidate.CoverURL != "https://example.test/large.jpg" {
		t.Fatalf("CoverURL = %q", candidate.CoverURL)
	}
	if candidate.Publisher != "Allen & Unwin" || candidate.Language != "en" {
		t.Fatalf("Publisher = %q, Language = %q", candidate.Publisher, candidate.Language)
	}
	if candidate.Genre != "Fantasy; Adventure" || candidate.Description == "" {
		t.Fatalf("Genre = %q, Description = %q", candidate.Genre, candidate.Description)
	}
}

func TestSearchAddsAPIKey(t *testing.T) {
	var gotKey string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotKey = r.URL.Query().Get("key")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"items":[]}`))
	}))
	defer server.Close()

	provider := New(
		WithBaseURL(server.URL),
		WithHTTPClient(server.Client()),
		WithAPIKey("test-key"),
	)

	if _, err := provider.Search(context.Background(), providers.SearchQuery{Title: "Book"}); err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if gotKey != "test-key" {
		t.Fatalf("key = %q, want test-key", gotKey)
	}
}

func TestGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/volumes/volume-1" {
			t.Fatalf("path = %q, want /volumes/volume-1", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"id": "volume-1",
			"volumeInfo": {
				"title": "Book",
				"authors": ["Author"]
			}
		}`))
	}))
	defer server.Close()

	got, err := New(WithBaseURL(server.URL), WithHTTPClient(server.Client())).Get(context.Background(), "volume-1")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if got.Title != "Book" {
		t.Fatalf("Title = %q", got.Title)
	}
}

func TestGetAddsAPIKey(t *testing.T) {
	var gotKey string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotKey = r.URL.Query().Get("key")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"volume-1","volumeInfo":{"title":"Book"}}`))
	}))
	defer server.Close()

	provider := New(
		WithBaseURL(server.URL),
		WithHTTPClient(server.Client()),
		WithAPIKey("test-key"),
	)

	if _, err := provider.Get(context.Background(), "volume-1"); err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if gotKey != "test-key" {
		t.Fatalf("key = %q, want test-key", gotKey)
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

func TestHelpers(t *testing.T) {
	if got := googleQuery(providers.SearchQuery{ISBN: "9780000000000"}); got != "isbn:9780000000000" {
		t.Fatalf("googleQuery() = %q", got)
	}
	if got := publishedYear("not-a-date"); got != 0 {
		t.Fatalf("publishedYear() = %d, want 0", got)
	}
	if got := bestImage(imageLinks{Thumbnail: "http://example.test/thumb.jpg"}); got != "https://example.test/thumb.jpg" {
		t.Fatalf("bestImage() = %q", got)
	}
}
