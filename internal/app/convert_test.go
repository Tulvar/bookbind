package app

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Tulvar/bookbind/internal/audio"
	"github.com/Tulvar/bookbind/internal/m4b"
	"github.com/Tulvar/bookbind/internal/metadata"
)

func TestConvertPlansDefaultOutput(t *testing.T) {
	dir := t.TempDir()
	inputPath := filepath.Join(dir, "book.mp3")
	if err := os.WriteFile(inputPath, []byte("test"), 0o644); err != nil {
		t.Fatalf("write test file: %v", err)
	}

	result, err := newTestApp().Convert(context.Background(), ConvertRequest{
		InputPath: inputPath,
		DryRun:    true,
	})
	if err != nil {
		t.Fatalf("Convert() error = %v", err)
	}

	if got, want := result.OutputPath, filepath.Join(dir, "book.m4b"); got != want {
		t.Fatalf("OutputPath = %q, want %q", got, want)
	}
	if !result.DryRun {
		t.Fatal("DryRun = false, want true")
	}
}

func TestConvertRejectsM4BInput(t *testing.T) {
	dir := t.TempDir()
	inputPath := filepath.Join(dir, "book.m4b")
	if err := os.WriteFile(inputPath, []byte("test"), 0o644); err != nil {
		t.Fatalf("write test file: %v", err)
	}

	_, err := newTestApp().Convert(context.Background(), ConvertRequest{
		InputPath: inputPath,
		DryRun:    true,
	})
	if err == nil || !strings.Contains(err.Error(), "only mp3 files") {
		t.Fatalf("Convert() error = %v, want mp3-only error", err)
	}
}

func TestConvertRejectsDirectoryContainingM4A(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"01.mp3", "02.m4a"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("test"), 0o644); err != nil {
			t.Fatalf("write test file: %v", err)
		}
	}

	_, err := newTestApp().Convert(context.Background(), ConvertRequest{
		InputPath: dir,
		DryRun:    true,
	})
	if err == nil || !strings.Contains(err.Error(), "only mp3 files") {
		t.Fatalf("Convert() error = %v, want mp3-only error", err)
	}
}

func TestConvertRejectsNonM4BOutput(t *testing.T) {
	dir := t.TempDir()
	inputPath := filepath.Join(dir, "book.mp3")
	if err := os.WriteFile(inputPath, []byte("test"), 0o644); err != nil {
		t.Fatalf("write test file: %v", err)
	}

	_, err := newTestApp().Convert(context.Background(), ConvertRequest{
		InputPath:  inputPath,
		OutputPath: filepath.Join(dir, "book.mp3"),
	})
	if err == nil {
		t.Fatal("Convert() error = nil, want error")
	}
}

func TestConvertRejectsExistingOutputWithoutOverwrite(t *testing.T) {
	dir := t.TempDir()
	inputPath := filepath.Join(dir, "book.mp3")
	outputPath := filepath.Join(dir, "book.m4b")
	if err := os.WriteFile(inputPath, []byte("test"), 0o644); err != nil {
		t.Fatalf("write input file: %v", err)
	}
	if err := os.WriteFile(outputPath, []byte("existing"), 0o644); err != nil {
		t.Fatalf("write output file: %v", err)
	}

	_, err := newTestApp().Convert(context.Background(), ConvertRequest{
		InputPath:  inputPath,
		OutputPath: outputPath,
	})
	if err == nil {
		t.Fatal("Convert() error = nil, want error")
	}
}

func TestConvertLoadsMetadataYAML(t *testing.T) {
	dir := t.TempDir()
	inputPath := filepath.Join(dir, "book.mp3")
	metadataPath := filepath.Join(dir, "bookbind.yaml")
	if err := os.WriteFile(inputPath, []byte("test"), 0o644); err != nil {
		t.Fatalf("write input file: %v", err)
	}
	if err := os.WriteFile(metadataPath, []byte(`title: "Night Watch"
author: "Sergey Lukyanenko"
`), 0o644); err != nil {
		t.Fatalf("write metadata file: %v", err)
	}

	result, err := newTestApp().Convert(context.Background(), ConvertRequest{
		InputPath:    inputPath,
		MetadataPath: metadataPath,
		DryRun:       true,
	})
	if err != nil {
		t.Fatalf("Convert() error = %v", err)
	}

	if got, want := result.Metadata.Title, "Night Watch"; got != want {
		t.Fatalf("Metadata.Title = %q, want %q", got, want)
	}
	if got := result.Metadata.NormalizedAuthors(); len(got) != 1 || got[0] != "Sergey Lukyanenko" {
		t.Fatalf("authors = %#v", got)
	}
}

func TestConvertUsesInlineMetadata(t *testing.T) {
	dir := t.TempDir()
	inputPath := filepath.Join(dir, "book.mp3")
	if err := os.WriteFile(inputPath, []byte("test"), 0o644); err != nil {
		t.Fatalf("write input file: %v", err)
	}

	result, err := newTestApp().Convert(context.Background(), ConvertRequest{
		InputPath: inputPath,
		Metadata:  metadata.Book{Title: "Inline Book"},
		DryRun:    true,
	})
	if err != nil {
		t.Fatalf("Convert() error = %v", err)
	}

	if got, want := result.Metadata.Title, "Inline Book"; got != want {
		t.Fatalf("Metadata.Title = %q, want %q", got, want)
	}
}

func TestPrepareConversionUsesDirectoryEmbeddedBookTags(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"01.mp3", "02.mp3"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("test"), 0o644); err != nil {
			t.Fatalf("write input file: %v", err)
		}
	}

	result, err := newTestAppWithProber(testProber{
		tags: audio.EmbeddedTags{
			Album:  "The Book",
			Artist: "The Author",
			Genre:  "Audiobook",
			Date:   "2022",
		},
	}).PrepareConversion(context.Background(), PrepareConversionRequest{InputPath: dir})
	if err != nil {
		t.Fatalf("PrepareConversion() error = %v", err)
	}

	if got, want := result.Metadata.Title, "The Book"; got != want {
		t.Fatalf("Title = %q, want %q", got, want)
	}
	if got, want := result.Metadata.Author, "The Author"; got != want {
		t.Fatalf("Author = %q, want %q", got, want)
	}
	if got, want := result.Metadata.PublishedYear, 2022; got != want {
		t.Fatalf("PublishedYear = %d, want %d", got, want)
	}
}

func TestPrepareConversionRejectsM4AInput(t *testing.T) {
	dir := t.TempDir()
	inputPath := filepath.Join(dir, "book.m4a")
	if err := os.WriteFile(inputPath, []byte("test"), 0o644); err != nil {
		t.Fatalf("write test file: %v", err)
	}

	_, err := newTestApp().PrepareConversion(context.Background(), PrepareConversionRequest{
		InputPath: inputPath,
	})
	if err == nil || !strings.Contains(err.Error(), "only mp3 files") {
		t.Fatalf("PrepareConversion() error = %v, want mp3-only error", err)
	}
}

func TestConvertUsesCLICover(t *testing.T) {
	dir := t.TempDir()
	inputPath := filepath.Join(dir, "book.mp3")
	coverPath := filepath.Join(dir, "cover.jpg")
	if err := os.WriteFile(inputPath, []byte("test"), 0o644); err != nil {
		t.Fatalf("write input file: %v", err)
	}
	if err := os.WriteFile(coverPath, []byte("cover"), 0o644); err != nil {
		t.Fatalf("write cover file: %v", err)
	}

	result, err := newTestApp().Convert(context.Background(), ConvertRequest{
		InputPath: inputPath,
		CoverPath: coverPath,
		DryRun:    true,
	})
	if err != nil {
		t.Fatalf("Convert() error = %v", err)
	}

	if got, want := result.CoverPath, coverPath; got != want {
		t.Fatalf("CoverPath = %q, want %q", got, want)
	}
}

func TestConvertResolvesRelativeMetadataCover(t *testing.T) {
	dir := t.TempDir()
	inputPath := filepath.Join(dir, "book.mp3")
	metadataPath := filepath.Join(dir, "bookbind.yaml")
	coverPath := filepath.Join(dir, "cover.jpg")
	if err := os.WriteFile(inputPath, []byte("test"), 0o644); err != nil {
		t.Fatalf("write input file: %v", err)
	}
	if err := os.WriteFile(coverPath, []byte("cover"), 0o644); err != nil {
		t.Fatalf("write cover file: %v", err)
	}
	if err := os.WriteFile(metadataPath, []byte(`title: "Book"
cover: "cover.jpg"
`), 0o644); err != nil {
		t.Fatalf("write metadata file: %v", err)
	}

	result, err := newTestApp().Convert(context.Background(), ConvertRequest{
		InputPath:    inputPath,
		MetadataPath: metadataPath,
		DryRun:       true,
	})
	if err != nil {
		t.Fatalf("Convert() error = %v", err)
	}

	if got, want := result.CoverPath, coverPath; got != want {
		t.Fatalf("CoverPath = %q, want %q", got, want)
	}
}

func TestConvertDownloadsRemoteMetadataCover(t *testing.T) {
	dir := t.TempDir()
	cachePath := filepath.Join(dir, "cache")
	previousCacheDir := cacheDir
	cacheDir = func() (string, error) { return cachePath, nil }
	t.Cleanup(func() { cacheDir = previousCacheDir })

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/jpeg")
		_, _ = w.Write([]byte("cover"))
	}))
	defer server.Close()

	inputPath := filepath.Join(dir, "book.mp3")
	metadataPath := filepath.Join(dir, "bookbind.yaml")
	if err := os.WriteFile(inputPath, []byte("test"), 0o644); err != nil {
		t.Fatalf("write input file: %v", err)
	}
	if err := os.WriteFile(metadataPath, []byte("title: \"Book\"\ncover: \""+server.URL+"/cover?id=book\"\n"), 0o644); err != nil {
		t.Fatalf("write metadata file: %v", err)
	}

	result, err := newTestApp().Convert(context.Background(), ConvertRequest{
		InputPath:    inputPath,
		MetadataPath: metadataPath,
		DryRun:       true,
	})
	if err != nil {
		t.Fatalf("Convert() error = %v", err)
	}

	if result.CoverPath == "" {
		t.Fatal("CoverPath is empty, want downloaded cover")
	}
	if _, err := os.Stat(result.CoverPath); err != nil {
		t.Fatalf("downloaded cover stat: %v", err)
	}
	if filepath.Dir(result.CoverPath) != filepath.Join(cachePath, "covers") {
		t.Fatalf("CoverPath = %q, want cache covers dir", result.CoverPath)
	}
}

func TestConvertRejectsUnsupportedCoverExtension(t *testing.T) {
	dir := t.TempDir()
	inputPath := filepath.Join(dir, "book.mp3")
	coverPath := filepath.Join(dir, "cover.gif")
	if err := os.WriteFile(inputPath, []byte("test"), 0o644); err != nil {
		t.Fatalf("write input file: %v", err)
	}
	if err := os.WriteFile(coverPath, []byte("cover"), 0o644); err != nil {
		t.Fatalf("write cover file: %v", err)
	}

	_, err := newTestApp().Convert(context.Background(), ConvertRequest{
		InputPath: inputPath,
		CoverPath: coverPath,
		DryRun:    true,
	})
	if err == nil {
		t.Fatal("Convert() error = nil, want error")
	}
}

func TestConvertAcceptsChapterEvery(t *testing.T) {
	dir := t.TempDir()
	inputPath := filepath.Join(dir, "book.mp3")
	if err := os.WriteFile(inputPath, []byte("test"), 0o644); err != nil {
		t.Fatalf("write input file: %v", err)
	}

	result, err := newTestApp().Convert(context.Background(), ConvertRequest{
		InputPath:    inputPath,
		ChapterEvery: "10m",
		DryRun:       true,
	})
	if err != nil {
		t.Fatalf("Convert() error = %v", err)
	}

	if got, want := result.ChapterEvery, "10m"; got != want {
		t.Fatalf("ChapterEvery = %q, want %q", got, want)
	}
}

func TestConvertRejectsInvalidChapterEvery(t *testing.T) {
	dir := t.TempDir()
	inputPath := filepath.Join(dir, "book.mp3")
	if err := os.WriteFile(inputPath, []byte("test"), 0o644); err != nil {
		t.Fatalf("write input file: %v", err)
	}

	_, err := newTestApp().Convert(context.Background(), ConvertRequest{
		InputPath:    inputPath,
		ChapterEvery: "soon",
		DryRun:       true,
	})
	if err == nil {
		t.Fatal("Convert() error = nil, want error")
	}
}

func newTestApp() *App {
	return newTestAppWithProber(testProber{})
}

func newTestAppWithProber(prober testProber) *App {
	builder := m4b.NewBuilder("ffmpeg")
	builder.Runner = testRunner{}
	return New(
		WithInspector(audio.NewInspector(audio.WithProber(prober))),
		WithBuilder(builder),
	)
}

type testProber struct {
	tags audio.EmbeddedTags
}

func (p testProber) Probe(context.Context, string) (audio.ProbeResult, error) {
	return audio.ProbeResult{
		Duration: 3 * time.Second,
		Codec:    "mp3",
		Bitrate:  128000,
		Channels: 2,
		Tags:     p.tags,
	}, nil
}

type testRunner struct{}

func (testRunner) Run(context.Context, string, ...string) error {
	return nil
}
