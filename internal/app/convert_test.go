package app

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
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

func TestConvertFinalMetadataOverridesCollectedValues(t *testing.T) {
	dir := t.TempDir()
	inputPath := filepath.Join(dir, "book.mp3")
	metadataPath := filepath.Join(dir, "bookbind.yaml")
	coverPath := filepath.Join(dir, "manual-cover.jpg")
	if err := os.WriteFile(inputPath, []byte("test"), 0o644); err != nil {
		t.Fatalf("write input file: %v", err)
	}
	if err := os.WriteFile(coverPath, []byte("cover"), 0o644); err != nil {
		t.Fatalf("write cover file: %v", err)
	}
	if err := os.WriteFile(metadataPath, []byte(`subtitle: "Provider Subtitle"
translators:
  - "Provider Translator"
publisher: "Provider Publisher"
`), 0o644); err != nil {
		t.Fatalf("write metadata file: %v", err)
	}

	manual := metadata.Book{
		Title:         "Восстание Персеполиса",
		Subtitle:      "Исправленный подзаголовок",
		Author:        "Джеймс С. А. Кори",
		Narrator:      "Всеволод Кузнецов",
		Translator:    "Галина Соловьева",
		Series:        "Пространство",
		SeriesIndex:   "7",
		Language:      "ru",
		Genre:         "Боевая фантастика",
		Description:   "Исправленное описание",
		Publisher:     "СОЮЗ",
		PublishedYear: 2021,
		Cover:         coverPath,
	}
	result, err := newTestAppWithProber(testProber{tags: audio.EmbeddedTags{
		Title:       "Пролог. Кортасар",
		Artist:      "Джеймс С. А. Кори",
		AlbumArtist: "Всеволод Кузнецов",
		Album:       "Восстание Персеполиса",
		Composer:    "Переводчик: Галина Соловьева",
		Genre:       "Embedded Genre",
		Date:        "2020",
		Comment:     "Embedded Description",
		Language:    "en",
	}}).Convert(context.Background(), ConvertRequest{
		InputPath:        inputPath,
		MetadataPath:     metadataPath,
		Metadata:         manual,
		MetadataOverride: true,
		DryRun:           true,
	})
	if err != nil {
		t.Fatalf("Convert() error = %v", err)
	}

	if !reflect.DeepEqual(result.Metadata, manual) {
		t.Fatalf("Metadata = %#v, want manual values %#v", result.Metadata, manual)
	}
}

func TestConvertInlineMetadataWithoutOverrideKeepsEmbeddedValues(t *testing.T) {
	dir := t.TempDir()
	inputPath := filepath.Join(dir, "book.mp3")
	if err := os.WriteFile(inputPath, []byte("test"), 0o644); err != nil {
		t.Fatalf("write input file: %v", err)
	}

	result, err := newTestAppWithProber(testProber{tags: audio.EmbeddedTags{
		Title:  "Embedded Title",
		Artist: "Embedded Author",
	}}).Convert(context.Background(), ConvertRequest{
		InputPath: inputPath,
		Metadata: metadata.Book{
			Title:  "Fallback Title",
			Author: "Fallback Author",
			Genre:  "Fallback Genre",
		},
		DryRun: true,
	})
	if err != nil {
		t.Fatalf("Convert() error = %v", err)
	}

	if got, want := result.Metadata.Title, "Embedded Title"; got != want {
		t.Fatalf("Metadata.Title = %q, want %q", got, want)
	}
	if got, want := result.Metadata.Author, "Embedded Author"; got != want {
		t.Fatalf("Metadata.Author = %q, want %q", got, want)
	}
	if got, want := result.Metadata.Genre, "Fallback Genre"; got != want {
		t.Fatalf("Metadata.Genre = %q, want %q", got, want)
	}
}

func TestPrepareConversionKeepsEmbeddedMetadataAndFillsMissingFields(t *testing.T) {
	dir := t.TempDir()
	inputPath := filepath.Join(dir, "Filename Title.mp3")
	if err := os.WriteFile(inputPath, []byte("test"), 0o644); err != nil {
		t.Fatalf("write input file: %v", err)
	}

	result, err := newTestAppWithProber(testProber{
		tags: audio.EmbeddedTags{
			Title:  "Embedded Title",
			Artist: "Embedded Author",
		},
	}).PrepareConversion(context.Background(), PrepareConversionRequest{
		InputPath: inputPath,
		Metadata: metadata.Book{
			Title:     "Internet Title",
			Author:    "Internet Author",
			Genre:     "Fantasy",
			Publisher: "Publisher",
		},
	})
	if err != nil {
		t.Fatalf("PrepareConversion() error = %v", err)
	}

	if result.Metadata.Title != "Embedded Title" || result.Metadata.Author != "Embedded Author" {
		t.Fatalf("Title = %q, Author = %q", result.Metadata.Title, result.Metadata.Author)
	}
	if result.Metadata.Genre != "Fantasy" || result.Metadata.Publisher != "Publisher" {
		t.Fatalf("Genre = %q, Publisher = %q", result.Metadata.Genre, result.Metadata.Publisher)
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

func TestPrepareConversionSeparatesEmbeddedAuthorAndNarrator(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"01.mp3", "02.mp3"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("test"), 0o644); err != nil {
			t.Fatalf("write input file: %v", err)
		}
	}

	result, err := newTestAppWithProber(testProber{
		tags: audio.EmbeddedTags{
			Album:       "Последний довод королей",
			Artist:      "Джо Аберкромби",
			AlbumArtist: "Читает Кирилл Головин",
		},
	}).PrepareConversion(context.Background(), PrepareConversionRequest{InputPath: dir})
	if err != nil {
		t.Fatalf("PrepareConversion() error = %v", err)
	}

	if got, want := result.Metadata.Author, "Джо Аберкромби"; got != want {
		t.Fatalf("Author = %q, want %q", got, want)
	}
	if got, want := result.Metadata.Narrator, "Кирилл Головин"; got != want {
		t.Fatalf("Narrator = %q, want %q", got, want)
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

func TestConvertDetectsLocalCover(t *testing.T) {
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
		DryRun:    true,
	})
	if err != nil {
		t.Fatalf("Convert() error = %v", err)
	}

	if got, want := result.CoverPath, coverPath; got != want {
		t.Fatalf("CoverPath = %q, want %q", got, want)
	}
	if !containsAdjacentArguments(result.Command, "-i", coverPath) {
		t.Fatalf("command does not include detected cover: %#v", result.Command)
	}
}

func TestConvertUsesEmbeddedCoverWhenNoExternalCoverExists(t *testing.T) {
	dir := t.TempDir()
	inputPath := filepath.Join(dir, "book.mp3")
	if err := os.WriteFile(inputPath, []byte("test"), 0o644); err != nil {
		t.Fatalf("write input file: %v", err)
	}

	result, err := newTestAppWithProber(testProber{hasAttachedPicture: true}).Convert(context.Background(), ConvertRequest{
		InputPath: inputPath,
		DryRun:    true,
	})
	if err != nil {
		t.Fatalf("Convert() error = %v", err)
	}

	if result.CoverPath != "" {
		t.Fatalf("CoverPath = %q for embedded cover, want empty external path", result.CoverPath)
	}
	if !containsAdjacentArguments(result.Command, "-map", "0:v:0") {
		t.Fatalf("command does not map embedded cover: %#v", result.Command)
	}
}

func TestConvertResolvesRelativeMetadataCover(t *testing.T) {
	dir := t.TempDir()
	inputPath := filepath.Join(dir, "book.mp3")
	metadataPath := filepath.Join(dir, "bookbind.yaml")
	coverPath := filepath.Join(dir, "metadata-cover.png")
	if err := os.WriteFile(inputPath, []byte("test"), 0o644); err != nil {
		t.Fatalf("write input file: %v", err)
	}
	if err := os.WriteFile(coverPath, []byte("cover"), 0o644); err != nil {
		t.Fatalf("write cover file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "cover.jpg"), []byte("automatic cover"), 0o644); err != nil {
		t.Fatalf("write automatic cover file: %v", err)
	}
	if err := os.WriteFile(metadataPath, []byte(`title: "Book"
cover: "metadata-cover.png"
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
	tags               audio.EmbeddedTags
	hasAttachedPicture bool
}

func (p testProber) Probe(context.Context, string) (audio.ProbeResult, error) {
	return audio.ProbeResult{
		Duration:           3 * time.Second,
		Codec:              "mp3",
		Bitrate:            128000,
		Channels:           2,
		NonAudioStreams:    boolInt(p.hasAttachedPicture),
		HasAttachedPicture: p.hasAttachedPicture,
		Tags:               p.tags,
	}, nil
}

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

func containsAdjacentArguments(command []string, want ...string) bool {
	if len(want) == 0 {
		return true
	}
	for start := 0; start+len(want) <= len(command); start++ {
		if reflect.DeepEqual(command[start:start+len(want)], want) {
			return true
		}
	}
	return false
}

type testRunner struct{}

func (testRunner) Run(context.Context, string, ...string) error {
	return nil
}
