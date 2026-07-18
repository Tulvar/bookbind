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
		Subtitle:      "The Other Side",
		Author:        "Sergey Lukyanenko",
		Narrator:      "Reader",
		Translator:    "Translator",
		Series:        "Watches",
		SeriesIndex:   "1",
		Genre:         "Fantasy",
		Description:   "Description",
		Publisher:     "Publisher",
		PublishedYear: 1998,
	}, nil)

	for _, want := range []string{
		"title=Night Watch: The Other Side",
		"disc_subtitle=The Other Side",
		"media_type=2",
		"artist=Sergey Lukyanenko",
		"composer=Reader",
		"album=Watches",
		"track=1",
		"genre=Fantasy",
		`description=Description\nNarrator: Reader\nTranslator: Translator\nPublisher: Publisher`,
		`synopsis=Description\nNarrator: Reader\nTranslator: Translator\nPublisher: Publisher`,
		"date=1998",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("metadata does not contain %q:\n%s", want, got)
		}
	}
	if strings.Contains(got, `album=Watches \#1`) {
		t.Fatalf("metadata embeds the series index in album:\n%s", got)
	}
}

func TestFFMetadataDocumentPreservesSeriesTrackTotal(t *testing.T) {
	got := FFMetadataDocument(metadata.Book{
		Series:      "Harry Potter",
		SeriesIndex: "1/12",
	}, nil)

	if !strings.Contains(got, "album=Harry Potter\n") {
		t.Fatalf("metadata does not contain the series album:\n%s", got)
	}
	if !strings.Contains(got, "track=1/12\n") {
		t.Fatalf("metadata does not contain the series track total:\n%s", got)
	}
}

func TestFFMetadataDocumentDoesNotDuplicateSubtitleInTitle(t *testing.T) {
	got := FFMetadataDocument(metadata.Book{
		Title:    "Night Watch: The Other Side",
		Subtitle: "The Other Side",
	}, nil)

	if strings.Contains(got, "title=Night Watch: The Other Side: The Other Side") {
		t.Fatalf("metadata duplicated subtitle:\n%s", got)
	}
}
