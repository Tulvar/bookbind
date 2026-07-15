package app

import (
	"testing"

	"github.com/Tulvar/bookbind/internal/audio"
)

func TestEmbeddedCredits(t *testing.T) {
	tests := []struct {
		name         string
		tags         audio.EmbeddedTags
		wantAuthor   string
		wantNarrator string
	}{
		{
			name: "russian narrator in album artist",
			tags: audio.EmbeddedTags{
				Artist:      "Джо Аберкромби",
				AlbumArtist: "Читает Кирилл Головин",
			},
			wantAuthor:   "Джо Аберкромби",
			wantNarrator: "Кирилл Головин",
		},
		{
			name: "ordinary album artist remains author",
			tags: audio.EmbeddedTags{
				Artist:      "Performer",
				AlbumArtist: "Book Author",
				Composer:    "Book Narrator",
			},
			wantAuthor:   "Book Author",
			wantNarrator: "Book Narrator",
		},
		{
			name: "marked artist becomes narrator",
			tags: audio.EmbeddedTags{
				Artist:      "Read by Jane Reader",
				AlbumArtist: "John Author",
			},
			wantAuthor:   "John Author",
			wantNarrator: "Jane Reader",
		},
		{
			name: "english narrator prefix with separator",
			tags: audio.EmbeddedTags{
				Artist:      "John Author",
				AlbumArtist: "Narrator: Jane Reader",
			},
			wantAuthor:   "John Author",
			wantNarrator: "Jane Reader",
		},
		{
			name: "unmarked album artist keeps existing priority",
			tags: audio.EmbeddedTags{
				Artist:      "Fallback Author",
				AlbumArtist: "Reader Guild",
			},
			wantAuthor: "Reader Guild",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			author, narrator := embeddedCredits(tt.tags)
			if author != tt.wantAuthor || narrator != tt.wantNarrator {
				t.Fatalf(
					"embeddedCredits() = author %q, narrator %q; want %q, %q",
					author,
					narrator,
					tt.wantAuthor,
					tt.wantNarrator,
				)
			}
		})
	}
}

func TestParseNarratorCreditRequiresNameAndBoundary(t *testing.T) {
	for _, value := range []string{"Читает", "Narrator", "NarratorName"} {
		if name, ok := parseNarratorCredit(value); ok || name != "" {
			t.Fatalf("parseNarratorCredit(%q) = %q, %v; want no match", value, name, ok)
		}
	}
}
