package app

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/Tulvar/bookbind/internal/metadata"
)

func TestTemplateMetadataWritesDefaultYAML(t *testing.T) {
	dir := t.TempDir()
	inputPath := filepath.Join(dir, "Night Watch.mp3")
	if err := os.WriteFile(inputPath, []byte("test"), 0o644); err != nil {
		t.Fatalf("write input file: %v", err)
	}

	result, err := newTestApp().TemplateMetadata(context.Background(), TemplateRequest{
		InputPath: inputPath,
	})
	if err != nil {
		t.Fatalf("TemplateMetadata() error = %v", err)
	}

	if got, want := result.OutputPath, filepath.Join(dir, "bookbind.yaml"); got != want {
		t.Fatalf("OutputPath = %q, want %q", got, want)
	}

	book, err := metadata.LoadYAML(result.OutputPath)
	if err != nil {
		t.Fatalf("LoadYAML() error = %v", err)
	}
	if got, want := book.Title, "Night Watch"; got != want {
		t.Fatalf("Title = %q, want %q", got, want)
	}
	if got, want := book.Language, "ru"; got != want {
		t.Fatalf("Language = %q, want %q", got, want)
	}
}

func TestTemplateMetadataUsesFilenameParser(t *testing.T) {
	dir := t.TempDir()
	inputPath := filepath.Join(dir, "Сергей Лукьяненко - Дозоры 01 - Ночной дозор.mp3")
	if err := os.WriteFile(inputPath, []byte("test"), 0o644); err != nil {
		t.Fatalf("write input file: %v", err)
	}

	result, err := newTestApp().TemplateMetadata(context.Background(), TemplateRequest{
		InputPath: inputPath,
	})
	if err != nil {
		t.Fatalf("TemplateMetadata() error = %v", err)
	}

	if got, want := result.Book.Author, "Сергей Лукьяненко"; got != want {
		t.Fatalf("Author = %q, want %q", got, want)
	}
	if got, want := result.Book.Series, "Дозоры"; got != want {
		t.Fatalf("Series = %q, want %q", got, want)
	}
	if got, want := result.Book.SeriesIndex, "01"; got != want {
		t.Fatalf("SeriesIndex = %q, want %q", got, want)
	}
	if got, want := result.Book.Title, "Ночной дозор"; got != want {
		t.Fatalf("Title = %q, want %q", got, want)
	}
}

func TestTemplateMetadataDetectsLocalCover(t *testing.T) {
	dir := t.TempDir()
	inputPath := filepath.Join(dir, "book.mp3")
	if err := os.WriteFile(inputPath, []byte("test"), 0o644); err != nil {
		t.Fatalf("write input file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "cover.jpg"), []byte("cover"), 0o644); err != nil {
		t.Fatalf("write cover file: %v", err)
	}

	result, err := newTestApp().TemplateMetadata(context.Background(), TemplateRequest{
		InputPath: inputPath,
	})
	if err != nil {
		t.Fatalf("TemplateMetadata() error = %v", err)
	}

	if got, want := result.Book.Cover, "cover.jpg"; got != want {
		t.Fatalf("Cover = %q, want %q", got, want)
	}
}

func TestTemplateMetadataDoesNotOverwriteByDefault(t *testing.T) {
	dir := t.TempDir()
	inputPath := filepath.Join(dir, "book.mp3")
	outputPath := filepath.Join(dir, "bookbind.yaml")
	if err := os.WriteFile(inputPath, []byte("test"), 0o644); err != nil {
		t.Fatalf("write input file: %v", err)
	}
	if err := os.WriteFile(outputPath, []byte("existing"), 0o644); err != nil {
		t.Fatalf("write existing template: %v", err)
	}

	_, err := newTestApp().TemplateMetadata(context.Background(), TemplateRequest{
		InputPath: inputPath,
	})
	if err == nil {
		t.Fatal("TemplateMetadata() error = nil, want error")
	}
}

func TestTemplateMetadataRejectsNonYAMLOutput(t *testing.T) {
	dir := t.TempDir()
	inputPath := filepath.Join(dir, "book.mp3")
	if err := os.WriteFile(inputPath, []byte("test"), 0o644); err != nil {
		t.Fatalf("write input file: %v", err)
	}

	_, err := newTestApp().TemplateMetadata(context.Background(), TemplateRequest{
		InputPath:  inputPath,
		OutputPath: filepath.Join(dir, "book.txt"),
	})
	if err == nil {
		t.Fatal("TemplateMetadata() error = nil, want error")
	}
}
