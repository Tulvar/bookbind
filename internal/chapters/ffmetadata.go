package chapters

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Tulvar/bookbind/internal/metadata"
)

func FFMetadata(chapters []Chapter) string {
	return FFMetadataDocument(metadata.Book{}, chapters)
}

func FFMetadataDocument(book metadata.Book, chapters []Chapter) string {
	var builder strings.Builder
	builder.WriteString(";FFMETADATA1\n")
	writeBookMetadata(&builder, book)
	for _, chapter := range chapters {
		builder.WriteString("\n[CHAPTER]\n")
		builder.WriteString("TIMEBASE=1/1000\n")
		fmt.Fprintf(&builder, "START=%d\n", millis(chapter.Start))
		fmt.Fprintf(&builder, "END=%d\n", millis(chapter.End))
		fmt.Fprintf(&builder, "title=%s\n", escapeValue(chapter.Title))
	}
	return builder.String()
}

func writeBookMetadata(builder *strings.Builder, book metadata.Book) {
	writeTag(builder, "media_type", "2")
	writeTag(builder, "title", book.Title)
	writeTag(builder, "artist", strings.Join(book.NormalizedAuthors(), "; "))
	writeTag(builder, "album_artist", strings.Join(book.NormalizedAuthors(), "; "))
	writeTag(builder, "composer", strings.Join(book.NormalizedNarrators(), "; "))
	writeTag(builder, "translator", strings.Join(book.NormalizedTranslators(), "; "))
	writeTag(builder, "album", albumTitle(book))
	writeTag(builder, "genre", book.Genre)
	writeTag(builder, "description", book.Description)
	writeTag(builder, "comment", book.Description)
	writeTag(builder, "publisher", book.Publisher)
	writeTag(builder, "language", book.Language)
	if book.PublishedYear > 0 {
		writeTag(builder, "date", strconv.Itoa(book.PublishedYear))
	}
}

func writeTag(builder *strings.Builder, key, value string) {
	value = strings.TrimSpace(value)
	if value == "" {
		return
	}
	fmt.Fprintf(builder, "%s=%s\n", key, escapeValue(value))
}

func albumTitle(book metadata.Book) string {
	if book.Series == "" {
		return ""
	}
	if book.SeriesIndex == "" {
		return book.Series
	}
	return book.Series + " #" + book.SeriesIndex
}

func millis(duration time.Duration) int64 {
	return duration.Round(time.Millisecond).Milliseconds()
}

func escapeValue(value string) string {
	value = strings.ReplaceAll(value, "\\", "\\\\")
	value = strings.ReplaceAll(value, "\n", "\\n")
	value = strings.ReplaceAll(value, "\r", "")
	value = strings.ReplaceAll(value, "=", "\\=")
	value = strings.ReplaceAll(value, ";", "\\;")
	value = strings.ReplaceAll(value, "#", "\\#")
	return value
}
