package app

import (
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/Tulvar/bookbind/internal/audio"
	"github.com/Tulvar/bookbind/internal/metadata"
)

func mergeEmbeddedTags(book metadata.Book, tags audio.EmbeddedTags) metadata.Book {
	if tags.Title != "" {
		book.Title = tags.Title
	}
	author, narrator := embeddedCredits(tags)
	if author != "" {
		book.Author = author
	}
	if tags.Album != "" {
		book.Series = tags.Album
	}
	if narrator != "" {
		book.Narrator = narrator
	}
	if tags.Genre != "" {
		book.Genre = tags.Genre
	}
	if tags.Comment != "" {
		book.Description = tags.Comment
	}
	if tags.Language != "" {
		book.Language = tags.Language
	}
	if year := parseYear(tags.Date); year > 0 {
		book.PublishedYear = year
	}
	return book
}

func embeddedCredits(tags audio.EmbeddedTags) (author, narrator string) {
	albumNarrator, albumArtistIsNarrator := parseNarratorCredit(tags.AlbumArtist)
	artistNarrator, artistIsNarrator := parseNarratorCredit(tags.Artist)

	if !albumArtistIsNarrator {
		author = strings.TrimSpace(tags.AlbumArtist)
	}
	if author == "" && !artistIsNarrator {
		author = strings.TrimSpace(tags.Artist)
	}

	narrator = firstNonEmpty(albumNarrator, artistNarrator, tags.Composer)
	return author, narrator
}

func parseNarratorCredit(value string) (string, bool) {
	value = strings.TrimSpace(value)
	lower := strings.ToLower(value)
	for _, prefix := range []string{
		"читает",
		"чтец",
		"диктор",
		"читає",
		"narrated by",
		"narrator",
		"read by",
	} {
		if !strings.HasPrefix(lower, prefix) {
			continue
		}
		remainder := value[len(prefix):]
		if remainder == "" {
			continue
		}
		first, _ := utf8.DecodeRuneInString(remainder)
		if !isCreditSeparator(first) {
			continue
		}
		name := strings.TrimLeftFunc(remainder, isCreditSeparator)
		name = strings.TrimSpace(name)
		if name != "" {
			return name, true
		}
	}
	return "", false
}

func isCreditSeparator(value rune) bool {
	return unicode.IsSpace(value) || strings.ContainsRune(":：-–—", value)
}

func parseYear(value string) int {
	value = strings.TrimSpace(value)
	if len(value) >= 4 {
		value = value[:4]
	}
	year, err := strconv.Atoi(value)
	if err != nil {
		return 0
	}
	return year
}
