package app

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Tulvar/bookbind/internal/filename"
	"github.com/Tulvar/bookbind/internal/metadata"
)

type TemplateRequest struct {
	InputPath  string
	OutputPath string
	Overwrite  bool
}

type TemplateResult struct {
	InputPath  string
	OutputPath string
	Book       metadata.Book
}

func (a *App) TemplateMetadata(ctx context.Context, req TemplateRequest) (TemplateResult, error) {
	input, err := a.inspector.Inspect(ctx, req.InputPath)
	if err != nil {
		return TemplateResult{}, err
	}

	outputPath := req.OutputPath
	if outputPath == "" {
		outputPath = filepath.Join(metadataTemplateDir(input.Path), "bookbind.yaml")
	}
	if err := validateTemplatePath(outputPath); err != nil {
		return TemplateResult{}, err
	}
	if err := ensureOutputWritable(outputPath, req.Overwrite); err != nil {
		return TemplateResult{}, err
	}

	book := filename.ParsePath(input.Path)
	if len(input.Files) == 1 {
		book = mergeEmbeddedTags(book, input.Files[0].Tags)
	}
	if book.Title == "" {
		book.Title = defaultTitle(input.Path)
	}
	if book.Language == "" {
		book.Language = "ru"
	}
	book.Cover = defaultCover(metadataTemplateDir(outputPath))

	data, err := metadata.MarshalTemplateYAML(book)
	if err != nil {
		return TemplateResult{}, err
	}
	if err := os.WriteFile(outputPath, data, 0o644); err != nil {
		return TemplateResult{}, err
	}

	return TemplateResult{
		InputPath:  input.Path,
		OutputPath: outputPath,
		Book:       book,
	}, nil
}

func metadataTemplateDir(path string) string {
	info, err := os.Stat(path)
	if err == nil && info.IsDir() {
		return path
	}
	return filepath.Dir(path)
}

func defaultTitle(inputPath string) string {
	base := filepath.Base(inputPath)
	if ext := filepath.Ext(base); ext != "" {
		base = base[:len(base)-len(ext)]
	}
	if base == "." || base == string(filepath.Separator) {
		return ""
	}
	return base
}

func defaultCover(dir string) string {
	for _, name := range []string{"cover.jpg", "cover.jpeg", "cover.png"} {
		path := filepath.Join(dir, name)
		info, err := os.Stat(path)
		if err == nil && !info.IsDir() {
			return name
		}
	}
	return ""
}

func validateTemplatePath(path string) error {
	if filepath.Ext(path) != ".yaml" && filepath.Ext(path) != ".yml" {
		return fmt.Errorf("template output must be .yaml or .yml")
	}
	return nil
}
