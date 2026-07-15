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
	input, err := a.inspectMP3Input(ctx, req.InputPath)
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
	book := inferEmbeddedMetadata(input)

	loaded, err := loadMetadata(metadataPath)
	if err != nil {
		return metadata.Book{}, err
	}
	book = fillMissingBook(book, loaded)
	book = fillMissingBook(book, inline)
	book = fillMissingBook(book, filename.ParsePath(input.Path))

	if book.Title == "" {
		book.Title = defaultTitle(input.Path)
	}
	if book.Language == "" {
		book.Language = "ru"
	}

	return book, nil
}

func inferEmbeddedMetadata(input audio.Input) metadata.Book {
	book := metadata.Book{}
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
	author, narrator := embeddedCredits(tags)
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

func fillMissingBook(base, fallback metadata.Book) metadata.Book {
	if base.Title == "" {
		base.Title = fallback.Title
	}
	if base.Subtitle == "" {
		base.Subtitle = fallback.Subtitle
	}
	if len(base.NormalizedAuthors()) == 0 {
		base.Authors = fallback.Authors
		base.Author = fallback.Author
	}
	if len(base.NormalizedNarrators()) == 0 {
		base.Narrators = fallback.Narrators
		base.Narrator = fallback.Narrator
	}
	if len(base.NormalizedTranslators()) == 0 {
		base.Translators = fallback.Translators
		base.Translator = fallback.Translator
	}
	if base.Series == "" {
		base.Series = fallback.Series
	}
	if base.SeriesIndex == "" {
		base.SeriesIndex = fallback.SeriesIndex
	}
	if base.Language == "" {
		base.Language = fallback.Language
	}
	if base.Genre == "" {
		base.Genre = fallback.Genre
	}
	if base.Description == "" {
		base.Description = fallback.Description
	}
	if base.Publisher == "" {
		base.Publisher = fallback.Publisher
	}
	if base.PublishedYear == 0 {
		base.PublishedYear = fallback.PublishedYear
	}
	if base.Cover == "" {
		base.Cover = fallback.Cover
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
