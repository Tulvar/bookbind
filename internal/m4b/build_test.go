package m4b

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/Tulvar/bookbind/internal/audio"
	"github.com/Tulvar/bookbind/internal/metadata"
)

func TestBuildDryRunSingleFilePlansCommand(t *testing.T) {
	builder := testBuilder()

	result, err := builder.Build(context.Background(), BuildRequest{
		Input: audio.Input{
			Files: []audio.File{{Path: "book.mp3"}},
		},
		OutputPath: "book.m4b",
		DryRun:     true,
	})
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	if !containsInOrder(result.Command, "ffmpeg", "-n", "-i", "book.mp3", "-i") {
		t.Fatalf("command does not contain metadata input: %#v", result.Command)
	}
	if !containsInOrder(result.Command, "-map", "0:a", "-map_metadata", "1") {
		t.Fatalf("command does not map metadata: %#v", result.Command)
	}
}

func TestBuildRunsFFmpeg(t *testing.T) {
	runner := &fakeRunner{}
	builder := testBuilder()
	builder.Runner = runner
	outputPath := filepath.Join(t.TempDir(), "book.m4b")

	result, err := builder.Build(context.Background(), BuildRequest{
		Input: audio.Input{
			Files: []audio.File{{Path: "book.mp3"}},
		},
		OutputPath: outputPath,
		Overwrite:  true,
	})
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	command := append([]string{runner.name}, runner.args...)
	if !containsInOrder(command, "ffmpeg", "-y", "-i", "book.mp3", "-i") {
		t.Fatalf("command does not contain expected inputs: %#v", command)
	}
	if !containsInOrder(command, "-c:a", "aac", "-b:a", "64k") {
		t.Fatalf("command does not contain conversion args: %#v", command)
	}
	if got := command[len(command)-1]; got == outputPath || filepath.Ext(got) != ".m4b" {
		t.Fatalf("ffmpeg output = %q, want temporary .m4b path", got)
	}
	if got := result.Command[len(result.Command)-1]; got != outputPath {
		t.Fatalf("reported output = %q, want %q", got, outputPath)
	}
	if data, readErr := os.ReadFile(outputPath); readErr != nil || string(data) != "complete" {
		t.Fatalf("published output = %q, error = %v", data, readErr)
	}
}

func TestBuildFailurePreservesExistingOutput(t *testing.T) {
	dir := t.TempDir()
	outputPath := filepath.Join(dir, "book.m4b")
	if err := os.WriteFile(outputPath, []byte("original"), 0o644); err != nil {
		t.Fatalf("write original output: %v", err)
	}

	builder := testBuilder()
	builder.Runner = runnerFunc(func(_ context.Context, _ string, args ...string) error {
		if err := os.WriteFile(args[len(args)-1], []byte("partial"), 0o644); err != nil {
			return err
		}
		return errors.New("conversion failed")
	})

	_, err := builder.Build(context.Background(), BuildRequest{
		Input:      audio.Input{Files: []audio.File{{Path: "book.mp3"}}},
		OutputPath: outputPath,
		Overwrite:  true,
	})
	if err == nil {
		t.Fatal("Build() error = nil, want error")
	}
	assertFileContents(t, outputPath, "original")
	assertNoTemporaryOutputs(t, dir)
}

func TestBuildPublishesOnlyAfterSuccessfulConversion(t *testing.T) {
	dir := t.TempDir()
	outputPath := filepath.Join(dir, "book.m4b")
	if err := os.WriteFile(outputPath, []byte("original"), 0o600); err != nil {
		t.Fatalf("write original output: %v", err)
	}

	var observedDuringConversion string
	builder := testBuilder()
	builder.Runner = runnerFunc(func(_ context.Context, _ string, args ...string) error {
		data, err := os.ReadFile(outputPath)
		if err != nil {
			return err
		}
		observedDuringConversion = string(data)
		return os.WriteFile(args[len(args)-1], []byte("complete"), 0o644)
	})

	_, err := builder.Build(context.Background(), BuildRequest{
		Input:      audio.Input{Files: []audio.File{{Path: "book.mp3"}}},
		OutputPath: outputPath,
		Overwrite:  true,
	})
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	if observedDuringConversion != "original" {
		t.Fatalf("output during conversion = %q, want original", observedDuringConversion)
	}
	assertFileContents(t, outputPath, "complete")
	info, err := os.Stat(outputPath)
	if err != nil {
		t.Fatalf("stat output: %v", err)
	}
	if got, want := info.Mode().Perm(), os.FileMode(0o600); got != want {
		t.Fatalf("output permissions = %o, want %o", got, want)
	}
	assertNoTemporaryOutputs(t, dir)
}

func TestBuildDoesNotReplaceOutputCreatedDuringConversion(t *testing.T) {
	dir := t.TempDir()
	outputPath := filepath.Join(dir, "book.m4b")
	builder := testBuilder()
	builder.Runner = runnerFunc(func(_ context.Context, _ string, args ...string) error {
		if err := os.WriteFile(outputPath, []byte("concurrent"), 0o644); err != nil {
			return err
		}
		return os.WriteFile(args[len(args)-1], []byte("complete"), 0o644)
	})

	_, err := builder.Build(context.Background(), BuildRequest{
		Input:      audio.Input{Files: []audio.File{{Path: "book.mp3"}}},
		OutputPath: outputPath,
	})
	if err == nil {
		t.Fatal("Build() error = nil, want publish conflict")
	}
	assertFileContents(t, outputPath, "concurrent")
	assertNoTemporaryOutputs(t, dir)
}

func TestNewBuilderResolvesFFmpegFromPath(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, testExecutableName("ffmpeg"))
	if err := os.WriteFile(path, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatalf("write ffmpeg: %v", err)
	}
	t.Setenv("PATH", dir)

	builder := NewBuilder("ffmpeg")
	if builder.FFmpegPath != path {
		t.Fatalf("FFmpegPath = %q, want %q", builder.FFmpegPath, path)
	}
}

func TestBuildSingleFileWithCoverMapsAttachedPicture(t *testing.T) {
	builder := testBuilder()

	result, err := builder.Build(context.Background(), BuildRequest{
		Input: audio.Input{
			Files: []audio.File{{Path: "book.mp3"}},
		},
		CoverPath:  "cover.jpg",
		OutputPath: "book.m4b",
		DryRun:     true,
	})
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	if !containsInOrder(result.Command, "-i", "cover.jpg", "-map", "2:v") {
		t.Fatalf("command does not map cover input: %#v", result.Command)
	}
	if !containsInOrder(result.Command, "-i", "book.mp3", "-i", "cover.jpg", "-map", "0:a") {
		t.Fatalf("command maps streams before declaring all inputs: %#v", result.Command)
	}
	if !containsInOrder(result.Command, "-c:v", "copy", "-disposition:v", "attached_pic") {
		t.Fatalf("command does not attach cover: %#v", result.Command)
	}
	if containsInOrder(result.Command, "-vn") {
		t.Fatalf("command should not disable video when cover exists: %#v", result.Command)
	}
}

func TestBuildSingleFileWithSyntheticChaptersMapsChapters(t *testing.T) {
	builder := testBuilder()

	result, err := builder.Build(context.Background(), BuildRequest{
		Input: audio.Input{
			Files: []audio.File{{Path: "book.mp3", Duration: 25 * time.Minute}},
		},
		OutputPath:   "book.m4b",
		ChapterEvery: 10 * time.Minute,
		DryRun:       true,
	})
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	if !containsInOrder(result.Command, "-map_metadata", "1", "-c:a", "aac", "-map_chapters", "1") {
		t.Fatalf("command does not map synthetic chapters: %#v", result.Command)
	}
}

func TestBuildWritesLanguageOnAudioTrack(t *testing.T) {
	builder := testBuilder()

	result, err := builder.Build(context.Background(), BuildRequest{
		Input: audio.Input{
			Files: []audio.File{{Path: "book.mp3"}},
		},
		Metadata:   metadata.Book{Language: "ru-RU"},
		OutputPath: "book.m4b",
		DryRun:     true,
	})
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	if !containsInOrder(result.Command, "-metadata:s:a:0", "language=rus") {
		t.Fatalf("command does not set audio language: %#v", result.Command)
	}
}

func TestBuildSingleFileUsesEmbeddedChapters(t *testing.T) {
	builder := testBuilder()

	result, err := builder.Build(context.Background(), BuildRequest{
		Input: audio.Input{
			Files: []audio.File{{
				Path:     "book.mp3",
				Duration: time.Minute,
				Chapters: []audio.Chapter{{
					Title: "Embedded Intro",
					Start: 0,
					End:   time.Minute,
				}},
			}},
		},
		OutputPath: "book.m4b",
		DryRun:     true,
	})
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	if !containsInOrder(result.Command, "-map_metadata", "1", "-c:a", "aac", "-map_chapters", "1") {
		t.Fatalf("command does not map embedded chapters: %#v", result.Command)
	}
}

func TestBuildSingleFileWithSyntheticChaptersRequiresDuration(t *testing.T) {
	builder := testBuilder()

	_, err := builder.Build(context.Background(), BuildRequest{
		Input: audio.Input{
			Files: []audio.File{{Path: "book.mp3"}},
		},
		OutputPath:   "book.m4b",
		ChapterEvery: 10 * time.Minute,
		DryRun:       true,
	})
	if err == nil {
		t.Fatal("Build() error = nil, want error")
	}
}

func TestBuildDirectoryUsesConcatDemuxer(t *testing.T) {
	builder := testBuilder()

	result, err := builder.Build(context.Background(), BuildRequest{
		Input: audio.Input{
			Files: []audio.File{
				{Path: filepath.Join("dir", "01.mp3"), Name: "01.mp3", Duration: time.Second},
				{Path: filepath.Join("dir", "02.mp3"), Name: "02.mp3", Duration: time.Second},
			},
		},
		OutputPath: "book.m4b",
		DryRun:     true,
	})
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	if got, want := filepath.Base(result.Command[0]), "ffmpeg"; got != want {
		t.Fatalf("command[0] = %q, want %q", got, want)
	}
	if !containsInOrder(result.Command, "-f", "concat", "-safe", "0", "-i") {
		t.Fatalf("command does not contain concat args in order: %#v", result.Command)
	}
	if !containsInOrder(result.Command, "-map_metadata", "1", "-map_chapters", "1") {
		t.Fatalf("command does not map ffmetadata chapters: %#v", result.Command)
	}
}

func TestBuildDirectoryWithCoverMapsAttachedPicture(t *testing.T) {
	builder := testBuilder()

	result, err := builder.Build(context.Background(), BuildRequest{
		Input: audio.Input{
			Files: []audio.File{
				{Path: filepath.Join("dir", "01.mp3"), Name: "01.mp3", Duration: time.Second},
				{Path: filepath.Join("dir", "02.mp3"), Name: "02.mp3", Duration: time.Second},
			},
		},
		CoverPath:  "cover.png",
		OutputPath: "book.m4b",
		DryRun:     true,
	})
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	if !containsInOrder(result.Command, "-i", "cover.png", "-map", "2:v") {
		t.Fatalf("command does not map cover input: %#v", result.Command)
	}
	if !containsInOrder(result.Command, "-i", "cover.png", "-map", "0:a") {
		t.Fatalf("command maps streams before declaring cover input: %#v", result.Command)
	}
	if !containsInOrder(result.Command, "-map_chapters", "1", "-c:a", "aac") {
		t.Fatalf("command does not keep chapters with cover: %#v", result.Command)
	}
}

func TestBuildDirectoryRejectsMissingDurations(t *testing.T) {
	builder := testBuilder()

	_, err := builder.Build(context.Background(), BuildRequest{
		Input: audio.Input{
			Files: []audio.File{
				{Path: filepath.Join("dir", "01.mp3"), Name: "01.mp3"},
				{Path: filepath.Join("dir", "02.mp3"), Name: "02.mp3"},
			},
		},
		OutputPath: "book.m4b",
		DryRun:     true,
	})
	if err == nil {
		t.Fatal("Build() error = nil, want error")
	}
}

func testBuilder() *Builder {
	return &Builder{
		FFmpegPath: "ffmpeg",
		Runner:     ExecRunner{},
	}
}

func testExecutableName(name string) string {
	if runtime.GOOS == "windows" {
		return name + ".exe"
	}
	return name
}

func assertCommand(t *testing.T, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("command len = %d, want %d\n got: %#v\nwant: %#v", len(got), len(want), got, want)
	}
	for i := range got {
		if got[i] != want[i] {
			t.Fatalf("command[%d] = %q, want %q\n got: %#v\nwant: %#v", i, got[i], want[i], got, want)
		}
	}
}

func containsInOrder(values []string, needles ...string) bool {
	if len(needles) == 0 {
		return true
	}
	next := 0
	for _, value := range values {
		if value == needles[next] {
			next++
			if next == len(needles) {
				return true
			}
		}
	}
	return false
}

type fakeRunner struct {
	name string
	args []string
}

func (r *fakeRunner) Run(_ context.Context, name string, args ...string) error {
	r.name = name
	r.args = args
	return os.WriteFile(args[len(args)-1], []byte("complete"), 0o644)
}

type runnerFunc func(context.Context, string, ...string) error

func (run runnerFunc) Run(ctx context.Context, name string, args ...string) error {
	return run(ctx, name, args...)
}

func assertFileContents(t *testing.T, path, want string) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	if got := string(data); got != want {
		t.Fatalf("contents of %s = %q, want %q", path, got, want)
	}
}

func assertNoTemporaryOutputs(t *testing.T, dir string) {
	t.Helper()
	matches, err := filepath.Glob(filepath.Join(dir, ".*.bookbind-*.m4b"))
	if err != nil {
		t.Fatalf("glob temporary outputs: %v", err)
	}
	if len(matches) != 0 {
		t.Fatalf("temporary outputs remain: %#v", matches)
	}
}
