package filename

import (
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Tulvar/bookbind/internal/metadata"
)

var (
	spaceRE       = regexp.MustCompile(`\s+`)
	seriesIndexRE = regexp.MustCompile(`^(.*?)(?:\s+|[._-]+)(\d{1,3})(?:\s*[-.]\s*|\s+)(.+)$`)
	leadingPartRE = regexp.MustCompile(`^\d{1,3}\s*[-.]\s*(.+)$`)
)

func ParsePath(path string) metadata.Book {
	base := cleanBase(filepath.Base(path))
	dir := filepath.Base(filepath.Dir(path))

	book := parseName(base)
	if book.Title == "" {
		book.Title = base
	}
	if book.Series == "" && looksSeriesDir(dir) {
		book.Series = cleanText(dir)
	}
	return book
}

func parseName(name string) metadata.Book {
	name = cleanText(name)
	if name == "" {
		return metadata.Book{}
	}

	if match := leadingPartRE.FindStringSubmatch(name); match != nil {
		return metadata.Book{Title: cleanText(match[1])}
	}

	parts := splitName(name)
	switch len(parts) {
	case 0:
		return metadata.Book{}
	case 1:
		return parseSeriesTitle(parts[0])
	case 2:
		right := parseSeriesTitle(parts[1])
		if right.Series != "" {
			right.Author = parts[0]
			return right
		}
		return metadata.Book{
			Author: parts[0],
			Title:  parts[1],
		}
	default:
		book := parseSeriesTitle(strings.Join(parts[1:], " - "))
		book.Author = parts[0]
		if book.Title == "" {
			book.Title = parts[len(parts)-1]
		}
		return book
	}
}

func parseSeriesTitle(value string) metadata.Book {
	value = cleanText(value)
	if match := seriesIndexRE.FindStringSubmatch(value); match != nil {
		return metadata.Book{
			Series:      cleanText(match[1]),
			SeriesIndex: cleanText(match[2]),
			Title:       cleanText(match[3]),
		}
	}
	return metadata.Book{Title: value}
}

func cleanBase(path string) string {
	ext := filepath.Ext(path)
	if ext != "" {
		path = strings.TrimSuffix(path, ext)
	}
	return cleanText(path)
}

func cleanText(value string) string {
	value = strings.ReplaceAll(value, "—", "-")
	value = strings.ReplaceAll(value, "–", "-")
	value = strings.ReplaceAll(value, "_", " ")
	value = strings.TrimSpace(value)
	value = spaceRE.ReplaceAllString(value, " ")
	return value
}

func splitName(value string) []string {
	rawParts := strings.Split(value, " - ")
	parts := make([]string, 0, len(rawParts))
	for _, part := range rawParts {
		part = cleanText(part)
		if part != "" {
			parts = append(parts, part)
		}
	}
	return parts
}

func looksSeriesDir(value string) bool {
	value = cleanText(value)
	if value == "" || value == "." || value == string(filepath.Separator) {
		return false
	}
	return !strings.Contains(value, ".")
}
