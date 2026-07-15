package app

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/Tulvar/bookbind/internal/metadata"
	"github.com/Tulvar/bookbind/internal/providers"
)

type ResolveMetadataRequest struct {
	Provider          string
	ID                string
	OutputPath        string
	Overwrite         bool
	GoogleBooksAPIKey string
}

type ResolveMetadataResult struct {
	OutputPath string
	Candidate  providers.Candidate
	Book       metadata.Book
}

type PreviewMetadataRequest struct {
	Provider          string
	ID                string
	GoogleBooksAPIKey string
}

type PreviewMetadataResult struct {
	Candidate providers.Candidate
	Book      metadata.Book
}

func (a *App) PreviewMetadata(ctx context.Context, req PreviewMetadataRequest) (PreviewMetadataResult, error) {
	candidate, err := a.getMetadataCandidate(ctx, req.Provider, req.ID, req.GoogleBooksAPIKey)
	if err != nil {
		return PreviewMetadataResult{}, err
	}

	return PreviewMetadataResult{
		Candidate: candidate,
		Book:      bookFromCandidate(candidate),
	}, nil
}

func (a *App) ResolveMetadata(ctx context.Context, req ResolveMetadataRequest) (ResolveMetadataResult, error) {
	candidate, err := a.getMetadataCandidate(ctx, req.Provider, req.ID, req.GoogleBooksAPIKey)
	if err != nil {
		return ResolveMetadataResult{}, err
	}

	outputPath := strings.TrimSpace(req.OutputPath)
	if outputPath == "" {
		outputPath = "bookbind.yaml"
	}
	if err := validateTemplatePath(outputPath); err != nil {
		return ResolveMetadataResult{}, err
	}
	if err := ensureOutputWritable(outputPath, req.Overwrite); err != nil {
		return ResolveMetadataResult{}, err
	}

	book := bookFromCandidate(candidate)
	data, err := metadata.MarshalTemplateYAML(book)
	if err != nil {
		return ResolveMetadataResult{}, err
	}
	if err := os.WriteFile(outputPath, data, 0o644); err != nil {
		return ResolveMetadataResult{}, err
	}

	return ResolveMetadataResult{
		OutputPath: outputPath,
		Candidate:  candidate,
		Book:       book,
	}, nil
}

func (a *App) getMetadataCandidate(ctx context.Context, provider, id, googleBooksAPIKey string) (providers.Candidate, error) {
	if a.providers == nil {
		return providers.Candidate{}, fmt.Errorf("metadata providers are not configured")
	}

	providerName := strings.TrimSpace(provider)
	if canonicalName, err := CanonicalProviderName(providerName); err == nil {
		providerName = canonicalName
	}
	if strings.TrimSpace(googleBooksAPIKey) != "" && providerName == "googlebooks" {
		registry, err := NewProviderRegistryWithConfig([]string{providerName}, ProviderConfig{
			GoogleBooksAPIKey: googleBooksAPIKey,
		})
		if err != nil {
			return providers.Candidate{}, err
		}
		return registry.Get(ctx, providerName, strings.TrimSpace(id))
	}

	candidate, err := a.providers.Get(ctx, providerName, strings.TrimSpace(id))
	if err != nil {
		return providers.Candidate{}, err
	}
	return candidate, nil
}

func bookFromCandidate(candidate providers.Candidate) metadata.Book {
	book := metadata.Book{
		Title:         candidate.Title,
		Subtitle:      candidate.Subtitle,
		Authors:       candidate.Authors,
		Narrators:     candidate.Narrators,
		Translators:   candidate.Translators,
		Series:        candidate.Series,
		SeriesIndex:   candidate.SeriesIndex,
		Language:      candidate.Language,
		Genre:         candidate.Genre,
		Description:   candidate.Description,
		Publisher:     candidate.Publisher,
		PublishedYear: candidate.Year,
		Cover:         candidate.CoverURL,
	}
	if len(book.Authors) > 0 {
		book.Author = strings.Join(book.Authors, ", ")
	}
	if len(book.Narrators) > 0 {
		book.Narrator = strings.Join(book.Narrators, ", ")
	}
	if len(book.Translators) > 0 {
		book.Translator = strings.Join(book.Translators, ", ")
	}
	return book
}
