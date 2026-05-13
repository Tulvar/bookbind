package app

import (
	"context"
	"strings"

	"github.com/Tulvar/bookbind/internal/audio"
	"github.com/Tulvar/bookbind/internal/filename"
	"github.com/Tulvar/bookbind/internal/metadata"
)

type PrepareConversionRequest struct {
	InputPath    string
	MetadataPath string
	Metadata     metadata.Book
}

type PrepareConversionResult struct {
	Input    audio.Input
	Metadata metadata.Book
	Missing  []string
}

func (a *App) PrepareConversion(ctx context.Context, req PrepareConversionRequest) (PrepareConversionResult, error) {
	input, err := a.inspector.Inspect(ctx, req.InputPath)
	if err != nil {
		return PrepareConversionResult{}, err
	}

	book, err := a.prepareMetadata(input, req.MetadataPath, req.Metadata)
	if err != nil {
		return PrepareConversionResult{}, err
	}

	return PrepareConversionResult{
		Input:    input,
		Metadata: book,
		Missing:  missingMetadataFields(book),
	}, nil
}

func (a *App) prepareMetadata(input audio.Input, metadataPath string, inline metadata.Book) (metadata.Book, error) {
	book := inferInputMetadata(input)

	loaded, err := loadMetadata(metadataPath)
	if err != nil {
		return metadata.Book{}, err
	}
	book = overlayBook(book, loaded)
	book = overlayBook(book, inline)

	if book.Title == "" {
		book.Title = defaultTitle(input.Path)
	}
	if book.Language == "" {
		book.Language = "ru"
	}

	return book, nil
}

func inferInputMetadata(input audio.Input) metadata.Book {
	book := filename.ParsePath(input.Path)
	if len(input.Files) == 0 {
		return book
	}
	if len(input.Files) == 1 {
		return mergeSingleFileTags(book, input.Files[0].Tags)
	}
	return mergeDirectoryTags(book, input.Files)
}

func mergeSingleFileTags(book metadata.Book, tags audio.EmbeddedTags) metadata.Book {
	if tags.Title != "" {
		book.Title = tags.Title
	}
	return mergeCommonTags(book, tags, false)
}

func mergeDirectoryTags(book metadata.Book, files []audio.File) metadata.Book {
	if album := commonTag(files, func(tags audio.EmbeddedTags) string { return tags.Album }); album != "" {
		book.Title = album
	}
	return mergeCommonTags(book, audio.EmbeddedTags{
		Artist:      commonTag(files, func(tags audio.EmbeddedTags) string { return tags.Artist }),
		AlbumArtist: commonTag(files, func(tags audio.EmbeddedTags) string { return tags.AlbumArtist }),
		Composer:    commonTag(files, func(tags audio.EmbeddedTags) string { return tags.Composer }),
		Genre:       commonTag(files, func(tags audio.EmbeddedTags) string { return tags.Genre }),
		Date:        commonTag(files, func(tags audio.EmbeddedTags) string { return tags.Date }),
		Comment:     commonTag(files, func(tags audio.EmbeddedTags) string { return tags.Comment }),
		Language:    commonTag(files, func(tags audio.EmbeddedTags) string { return tags.Language }),
	}, true)
}

func mergeCommonTags(book metadata.Book, tags audio.EmbeddedTags, albumIsTitle bool) metadata.Book {
	author := firstNonEmpty(tags.AlbumArtist, tags.Artist)
	if author != "" {
		book.Author = author
	}
	if tags.Album != "" {
		if albumIsTitle {
			book.Title = tags.Album
		} else if book.Series == "" {
			book.Series = tags.Album
		}
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

func commonTag(files []audio.File, value func(audio.EmbeddedTags) string) string {
	var common string
	for _, file := range files {
		current := strings.TrimSpace(value(file.Tags))
		if current == "" {
			continue
		}
		if common == "" {
			common = current
			continue
		}
		if !strings.EqualFold(common, current) {
			return ""
		}
	}
	return common
}

func overlayBook(base, override metadata.Book) metadata.Book {
	if override.Title != "" {
		base.Title = override.Title
	}
	if override.Subtitle != "" {
		base.Subtitle = override.Subtitle
	}
	if len(override.Authors) > 0 {
		base.Authors = override.Authors
		base.Author = ""
	} else if override.Author != "" {
		base.Author = override.Author
		base.Authors = nil
	}
	if len(override.Narrators) > 0 {
		base.Narrators = override.Narrators
		base.Narrator = ""
	} else if override.Narrator != "" {
		base.Narrator = override.Narrator
		base.Narrators = nil
	}
	if len(override.Translators) > 0 {
		base.Translators = override.Translators
		base.Translator = ""
	} else if override.Translator != "" {
		base.Translator = override.Translator
		base.Translators = nil
	}
	if override.Series != "" {
		base.Series = override.Series
	}
	if override.SeriesIndex != "" {
		base.SeriesIndex = override.SeriesIndex
	}
	if override.Language != "" {
		base.Language = override.Language
	}
	if override.Genre != "" {
		base.Genre = override.Genre
	}
	if override.Description != "" {
		base.Description = override.Description
	}
	if override.Publisher != "" {
		base.Publisher = override.Publisher
	}
	if override.PublishedYear > 0 {
		base.PublishedYear = override.PublishedYear
	}
	if override.Cover != "" {
		base.Cover = override.Cover
	}
	return base
}

func missingMetadataFields(book metadata.Book) []string {
	var missing []string
	if strings.TrimSpace(book.Title) == "" {
		missing = append(missing, "title")
	}
	if len(book.NormalizedAuthors()) == 0 {
		missing = append(missing, "author")
	}
	if len(book.NormalizedNarrators()) == 0 {
		missing = append(missing, "narrator")
	}
	if len(book.NormalizedTranslators()) == 0 {
		missing = append(missing, "translator")
	}
	if strings.TrimSpace(book.Language) == "" {
		missing = append(missing, "language")
	}
	if strings.TrimSpace(book.Cover) == "" {
		missing = append(missing, "cover")
	}
	return missing
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
