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
	writeTag(builder, "title", displayTitle(book))
	writeTag(builder, "disc_subtitle", book.Subtitle)
	writeTag(builder, "artist", strings.Join(book.NormalizedAuthors(), "; "))
	writeTag(builder, "album_artist", strings.Join(book.NormalizedAuthors(), "; "))
	writeTag(builder, "composer", strings.Join(book.NormalizedNarrators(), "; "))
	writeTag(builder, "album", book.Series)
	writeTag(builder, "track", book.SeriesIndex)
	writeTag(builder, "genre", book.Genre)
	description := compatibleDescription(book)
	writeTag(builder, "description", description)
	writeTag(builder, "synopsis", description)
	writeTag(builder, "comment", description)
	if book.PublishedYear > 0 {
		writeTag(builder, "date", strconv.Itoa(book.PublishedYear))
	}
}

func displayTitle(book metadata.Book) string {
	title := strings.TrimSpace(book.Title)
	subtitle := strings.TrimSpace(book.Subtitle)
	if title == "" {
		return subtitle
	}
	normalizedTitle := strings.ToLower(title)
	normalizedSubtitle := strings.ToLower(subtitle)
	if subtitle == "" || normalizedTitle == normalizedSubtitle || strings.HasSuffix(normalizedTitle, ": "+normalizedSubtitle) {
		return title
	}
	return title + ": " + subtitle
}

func compatibleDescription(book metadata.Book) string {
	parts := make([]string, 0, 4)
	if description := strings.TrimSpace(book.Description); description != "" {
		parts = append(parts, description)
	}
	if narrators := strings.Join(book.NormalizedNarrators(), "; "); narrators != "" {
		parts = append(parts, "Narrator: "+narrators)
	}
	if translators := strings.Join(book.NormalizedTranslators(), "; "); translators != "" {
		parts = append(parts, "Translator: "+translators)
	}
	if publisher := strings.TrimSpace(book.Publisher); publisher != "" {
		parts = append(parts, "Publisher: "+publisher)
	}
	return strings.Join(parts, "\n")
}

func writeTag(builder *strings.Builder, key, value string) {
	value = strings.TrimSpace(value)
	if value == "" {
		return
	}
	fmt.Fprintf(builder, "%s=%s\n", key, escapeValue(value))
}

func millis(duration time.Duration) int64 {
	return duration.Round(time.Millisecond).Milliseconds()
}

func escapeValue(value string) string {
	value = strings.ReplaceAll(value, "\\", "\\\\")
	value = strings.ReplaceAll(value, "\r", "")
	value = strings.ReplaceAll(value, "\n", "\\\n")
	value = strings.ReplaceAll(value, "=", "\\=")
	value = strings.ReplaceAll(value, ";", "\\;")
	value = strings.ReplaceAll(value, "#", "\\#")
	return value
}
