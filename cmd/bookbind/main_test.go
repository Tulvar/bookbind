package main

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/Tulvar/bookbind/internal/audio"
	"github.com/Tulvar/bookbind/internal/metadata"
	"github.com/Tulvar/bookbind/internal/providers"
)

func TestReorderFlagArgsAllowsFlagsAfterPositional(t *testing.T) {
	got := reorderFlagArgs(
		[]string{"book.mp3", "--output", "book.m4b", "--dry-run"},
		map[string]bool{
			"output":  true,
			"dry-run": false,
		},
	)
	want := []string{"--output", "book.m4b", "--dry-run", "book.mp3"}

	if len(got) != len(want) {
		t.Fatalf("len = %d, want %d: %#v", len(got), len(want), got)
	}
	for i := range got {
		if got[i] != want[i] {
			t.Fatalf("got[%d] = %q, want %q: %#v", i, got[i], want[i], got)
		}
	}
}

func TestReorderFlagArgsKeepsEqualsFlagTogether(t *testing.T) {
	got := reorderFlagArgs(
		[]string{"book.mp3", "--output=book.m4b"},
		map[string]bool{"output": true},
	)
	want := []string{"--output=book.m4b", "book.mp3"}

	if len(got) != len(want) {
		t.Fatalf("len = %d, want %d: %#v", len(got), len(want), got)
	}
	for i := range got {
		if got[i] != want[i] {
			t.Fatalf("got[%d] = %q, want %q: %#v", i, got[i], want[i], got)
		}
	}
}

func TestReorderFlagArgsAllowsTemplateOutputAfterPositional(t *testing.T) {
	got := reorderFlagArgs(
		[]string{"book.mp3", "--output", "custom.yaml"},
		map[string]bool{"output": true},
	)
	want := []string{"--output", "custom.yaml", "book.mp3"}

	if len(got) != len(want) {
		t.Fatalf("len = %d, want %d: %#v", len(got), len(want), got)
	}
	for i := range got {
		if got[i] != want[i] {
			t.Fatalf("got[%d] = %q, want %q: %#v", i, got[i], want[i], got)
		}
	}
}

func TestReorderFlagArgsAllowsChapterEveryAfterPositional(t *testing.T) {
	got := reorderFlagArgs(
		[]string{"book.mp3", "--chapter-every", "10m"},
		map[string]bool{"chapter-every": true},
	)
	want := []string{"--chapter-every", "10m", "book.mp3"}

	if len(got) != len(want) {
		t.Fatalf("len = %d, want %d: %#v", len(got), len(want), got)
	}
	for i := range got {
		if got[i] != want[i] {
			t.Fatalf("got[%d] = %q, want %q: %#v", i, got[i], want[i], got)
		}
	}
}

func TestReorderFlagArgsAllowsSearchFlags(t *testing.T) {
	got := reorderFlagArgs(
		[]string{"--title", "Ночной дозор", "--author", "Лукьяненко", "--provider", "openlibrary,googlebooks", "--select", "1", "--output", "bookbind.yaml", "--overwrite"},
		map[string]bool{"title": true, "author": true, "provider": true, "select": true, "output": true, "overwrite": false},
	)
	want := []string{"--title", "Ночной дозор", "--author", "Лукьяненко", "--provider", "openlibrary,googlebooks", "--select", "1", "--output", "bookbind.yaml", "--overwrite"}

	if len(got) != len(want) {
		t.Fatalf("len = %d, want %d: %#v", len(got), len(want), got)
	}
	for i := range got {
		if got[i] != want[i] {
			t.Fatalf("got[%d] = %q, want %q: %#v", i, got[i], want[i], got)
		}
	}
}

func TestSelectCandidate(t *testing.T) {
	got, err := selectCandidate([]providers.Candidate{
		{ID: "first"},
		{ID: "second"},
	}, 2)
	if err != nil {
		t.Fatalf("selectCandidate() error = %v", err)
	}

	if got.ID != "second" {
		t.Fatalf("ID = %q", got.ID)
	}
}

func TestSelectCandidateRejectsOutOfRange(t *testing.T) {
	_, err := selectCandidate([]providers.Candidate{{ID: "first"}}, 2)
	if err == nil {
		t.Fatal("selectCandidate() error = nil, want error")
	}
}

func TestSplitProviderList(t *testing.T) {
	got := splitProviderList(" openlibrary, googlebooks,, ")
	want := []string{"openlibrary", "googlebooks"}

	if len(got) != len(want) {
		t.Fatalf("len = %d, want %d: %#v", len(got), len(want), got)
	}
	for i := range got {
		if got[i] != want[i] {
			t.Fatalf("got[%d] = %q, want %q: %#v", i, got[i], want[i], got)
		}
	}
}

func TestRunProviders(t *testing.T) {
	var buffer bytes.Buffer

	if err := runProviders(&buffer); err != nil {
		t.Fatalf("runProviders() error = %v", err)
	}

	got := buffer.String()
	for _, want := range []string{
		"openlibrary\tenabled",
		"googlebooks\tenabled",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("output does not contain %q:\n%s", want, got)
		}
	}
}

func TestReorderFlagArgsAllowsMetadataFlags(t *testing.T) {
	got := reorderFlagArgs(
		[]string{"--provider", "googlebooks", "--id", "abc", "--preview", "--output", "bookbind.yaml", "--overwrite"},
		map[string]bool{"provider": true, "id": true, "output": true, "preview": false, "overwrite": false},
	)
	want := []string{"--provider", "googlebooks", "--id", "abc", "--preview", "--output", "bookbind.yaml", "--overwrite"}

	if len(got) != len(want) {
		t.Fatalf("len = %d, want %d: %#v", len(got), len(want), got)
	}
	for i := range got {
		if got[i] != want[i] {
			t.Fatalf("got[%d] = %q, want %q: %#v", i, got[i], want[i], got)
		}
	}
}

func TestPrintMetadataDetails(t *testing.T) {
	var buffer bytes.Buffer

	printMetadataDetails(&buffer, providers.Candidate{
		Provider:   "googlebooks",
		ID:         "abc",
		Confidence: 0.95,
	}, metadata.Book{
		Title:         "Night Watch",
		Author:        "Sergey Lukyanenko",
		Narrator:      "Reader",
		Series:        "Watches",
		SeriesIndex:   "1",
		PublishedYear: 1998,
		Cover:         "https://example.test/cover.jpg",
	})

	got := buffer.String()
	for _, want := range []string{
		"Provider: googlebooks",
		"ID: abc",
		"Title: Night Watch",
		"Author: Sergey Lukyanenko",
		"Narrator: Reader",
		"Series: Watches #1",
		"Published year: 1998",
		"Cover: https://example.test/cover.jpg",
		"Confidence: 0.95",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("output does not contain %q:\n%s", want, got)
		}
	}
}

func TestPrintSearchCandidatesUsesTable(t *testing.T) {
	var buffer bytes.Buffer

	err := printSearchCandidates(&buffer, []providers.Candidate{
		{
			Provider:   "googlebooks",
			ID:         "abc",
			Title:      "Night Watch",
			Authors:    []string{"Sergey Lukyanenko"},
			Year:       1998,
			Confidence: 0.95,
		},
	})
	if err != nil {
		t.Fatalf("printSearchCandidates() error = %v", err)
	}

	got := buffer.String()
	for _, want := range []string{
		"Candidates: 1",
		"#  Provider",
		"1  googlebooks",
		"abc",
		"Night Watch",
		"Sergey Lukyanenko",
		"1998",
		"0.95",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("output does not contain %q:\n%s", want, got)
		}
	}
}

func TestPrintSearchCandidatesHandlesEmptyResult(t *testing.T) {
	var buffer bytes.Buffer

	if err := printSearchCandidates(&buffer, nil); err != nil {
		t.Fatalf("printSearchCandidates() error = %v", err)
	}

	got := buffer.String()
	if got != "Candidates: 0\n" {
		t.Fatalf("output = %q", got)
	}
}

func TestPrintEmbeddedTags(t *testing.T) {
	var buffer bytes.Buffer

	printEmbeddedTags(&buffer, audio.EmbeddedTags{
		Title:  "Night Watch",
		Artist: "Sergey Lukyanenko",
	})

	got := buffer.String()
	for _, want := range []string{
		"Embedded metadata:",
		"title: Night Watch",
		"artist: Sergey Lukyanenko",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("output does not contain %q:\n%s", want, got)
		}
	}
}

func TestPrintChapters(t *testing.T) {
	var buffer bytes.Buffer

	printChapters(&buffer, []audio.Chapter{
		{Title: "Intro", Start: 0, End: time.Minute},
	})

	got := buffer.String()
	for _, want := range []string{
		"Chapters: 1",
		"Intro (00:00.000 - 01:00.000)",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("output does not contain %q:\n%s", want, got)
		}
	}
}
