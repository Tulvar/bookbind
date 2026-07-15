package fantlab

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
		_, _ = w.Write([]byte(`[
			{
				"work_id": "42",
				"rusname": "Последний довод королей",
				"all_autor_rusname": "Джо Аберкромби",
				"year": "2008",
				"pic_edition_id": "12345"
			}
		]`))
	}))
	defer server.Close()

	got, err := New(WithBaseURL(server.URL), WithHTTPClient(server.Client())).
		Search(context.Background(), providers.SearchQuery{
			Title:  "Последний довод королей",
			Author: "Аберкромби",
		})
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if gotPath != "/search-works" {
		t.Fatalf("path = %q, want /search-works", gotPath)
	}
	if !strings.Contains(gotQuery, "q=") || !strings.Contains(gotQuery, "onlymatches=1") {
		t.Fatalf("query = %q", gotQuery)
	}

	if len(got) != 1 {
		t.Fatalf("len = %d, want 1", len(got))
	}
	candidate := got[0]
	if candidate.Provider != "fantlab" {
		t.Fatalf("Provider = %q, want fantlab", candidate.Provider)
	}
	if candidate.ID != "42" {
		t.Fatalf("ID = %q, want 42", candidate.ID)
	}
	if candidate.Title != "Последний довод королей" {
		t.Fatalf("Title = %q", candidate.Title)
	}
	if candidate.Authors[0] != "Джо Аберкромби" {
		t.Fatalf("Author = %q", candidate.Authors[0])
	}
	if candidate.Year != 2008 {
		t.Fatalf("Year = %d", candidate.Year)
	}
	if candidate.CoverURL != "https://fantlab.ru/images/editions/big/12345" {
		t.Fatalf("CoverURL = %q", candidate.CoverURL)
	}
}

func TestGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/work/42" {
			t.Fatalf("path = %q, want /work/42", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"work_id": 42,
			"work_name": "Последний довод королей",
			"work_year": 2008,
			"work_description": "<p>Завершение трилогии.</p>",
			"lang_code": "ru",
			"image": "//fantlab.ru/images/work/42",
			"authors": [{"name": "Джо Аберкромби"}]
		}`))
	}))
	defer server.Close()

	got, err := New(WithBaseURL(server.URL), WithHTTPClient(server.Client())).
		Get(context.Background(), "42")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if got.Title != "Последний довод королей" {
		t.Fatalf("Title = %q", got.Title)
	}
	if got.CoverURL != "https://fantlab.ru/images/work/42" {
		t.Fatalf("CoverURL = %q", got.CoverURL)
	}
	if got.Description != "Завершение трилогии." || got.Language != "ru" {
		t.Fatalf("Description = %q, Language = %q", got.Description, got.Language)
	}
}

func TestSearchRejectsEmptyQuery(t *testing.T) {
	_, err := New().Search(context.Background(), providers.SearchQuery{})
	if err == nil {
		t.Fatal("Search() error = nil, want error")
	}
}
