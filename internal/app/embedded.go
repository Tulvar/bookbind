package app

import (
	"strconv"
	"strings"

	"github.com/Tulvar/bookbind/internal/audio"
	"github.com/Tulvar/bookbind/internal/metadata"
)

func mergeEmbeddedTags(book metadata.Book, tags audio.EmbeddedTags) metadata.Book {
	if tags.Title != "" {
		book.Title = tags.Title
	}
	if tags.Artist != "" {
		book.Author = tags.Artist
	}
	if tags.Album != "" {
		book.Series = tags.Album
	}
	if tags.Composer != "" {
		book.Narrator = tags.Composer
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
