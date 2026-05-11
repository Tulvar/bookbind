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
	Provider   string
	ID         string
	OutputPath string
	Overwrite  bool
}

type ResolveMetadataResult struct {
	OutputPath string
	Candidate  providers.Candidate
	Book       metadata.Book
}

func (a *App) ResolveMetadata(ctx context.Context, req ResolveMetadataRequest) (ResolveMetadataResult, error) {
	if a.providers == nil {
		return ResolveMetadataResult{}, fmt.Errorf("metadata providers are not configured")
	}

	providerName := strings.TrimSpace(req.Provider)
	if canonicalName, err := CanonicalProviderName(providerName); err == nil {
		providerName = canonicalName
	}

	candidate, err := a.providers.Get(ctx, providerName, strings.TrimSpace(req.ID))
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

func bookFromCandidate(candidate providers.Candidate) metadata.Book {
	book := metadata.Book{
		Title:         candidate.Title,
		Authors:       candidate.Authors,
		Narrators:     candidate.Narrators,
		Series:        candidate.Series,
		SeriesIndex:   candidate.SeriesIndex,
		PublishedYear: candidate.Year,
		Cover:         candidate.CoverURL,
	}
	if len(book.Authors) > 0 {
		book.Author = strings.Join(book.Authors, ", ")
	}
	if len(book.Narrators) > 0 {
		book.Narrator = strings.Join(book.Narrators, ", ")
	}
	return book
}
