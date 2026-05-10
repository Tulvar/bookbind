package chapters

import (
	"strings"
	"testing"
	"time"

	"github.com/Tulvar/bookbind/internal/metadata"
)

func TestFFMetadataWritesChapters(t *testing.T) {
	got := FFMetadata([]Chapter{
		{Title: "Chapter 1", Start: 0, End: 30 * time.Second},
		{Title: "Chapter 2", Start: 30 * time.Second, End: 75 * time.Second},
	})

	for _, want := range []string{
		";FFMETADATA1",
		"TIMEBASE=1/1000",
		"START=0",
		"END=30000",
		"title=Chapter 1",
		"START=30000",
		"END=75000",
		"title=Chapter 2",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("metadata does not contain %q:\n%s", want, got)
		}
	}
}

func TestFFMetadataEscapesSpecialValues(t *testing.T) {
	got := FFMetadata([]Chapter{
		{Title: "A=B;C#D", Start: 0, End: time.Second},
	})

	if !strings.Contains(got, `title=A\=B\;C\#D`) {
		t.Fatalf("metadata title was not escaped:\n%s", got)
	}
}

func TestFFMetadataDocumentWritesBookTags(t *testing.T) {
	got := FFMetadataDocument(metadata.Book{
		Title:         "Night Watch",
		Author:        "Sergey Lukyanenko",
		Narrator:      "Reader",
		Series:        "Watches",
		SeriesIndex:   "1",
		Genre:         "Fantasy",
		PublishedYear: 1998,
	}, nil)

	for _, want := range []string{
		"title=Night Watch",
		"artist=Sergey Lukyanenko",
		"composer=Reader",
		`album=Watches \#1`,
		"genre=Fantasy",
		"date=1998",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("metadata does not contain %q:\n%s", want, got)
		}
	}
}
