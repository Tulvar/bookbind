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
	originalInfo, statErr := os.Stat(outputPath)
	if statErr != nil {
		t.Fatalf("stat original output: %v", statErr)
	}
	originalMode := originalInfo.Mode().Perm()

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
	if got, want := info.Mode().Perm(), originalMode; got != want {
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

func TestBuildDirectoryUsesConcatDemuxerForCompatibleStreams(t *testing.T) {
	builder := testBuilder()

	result, err := builder.Build(context.Background(), BuildRequest{
		Input: audio.Input{
			Files: []audio.File{
				compatibleMP3(filepath.Join("dir", "01.mp3"), time.Second),
				compatibleMP3(filepath.Join("dir", "02.mp3"), time.Second),
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

func TestBuildDirectoryUsesConcatFilterForIncompatibleStreams(t *testing.T) {
	builder := testBuilder()
	first := compatibleMP3(filepath.Join("dir", "01.mp3"), time.Second)
	second := compatibleMP3(filepath.Join("dir", "02.mp3"), time.Second)
	second.SampleRate = 48000

	result, err := builder.Build(context.Background(), BuildRequest{
		Input:      audio.Input{Files: []audio.File{first, second}},
		OutputPath: "book.m4b",
		DryRun:     true,
	})
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	if containsInOrder(result.Command, "-f", "concat") {
		t.Fatalf("command uses concat demuxer for incompatible streams: %#v", result.Command)
	}
	if !containsInOrder(result.Command,
		"-i", first.Path,
		"-i", second.Path,
		"-i",
		"-filter_complex",
	) {
		t.Fatalf("command does not open audio files separately: %#v", result.Command)
	}
	wantFilter := "[0:a:0]asetpts=PTS-STARTPTS[bookbind_a0];" +
		"[1:a:0]asetpts=PTS-STARTPTS[bookbind_a1];" +
		"[bookbind_a0][bookbind_a1]concat=n=2:v=0:a=1[bookbind_audio]"
	if !containsInOrder(result.Command, "-filter_complex", wantFilter, "-map", "[bookbind_audio]") {
		t.Fatalf("command does not normalize and concatenate audio: %#v", result.Command)
	}
	if !containsInOrder(result.Command, "-map_metadata", "2", "-map_chapters", "2") {
		t.Fatalf("command does not map fallback metadata input: %#v", result.Command)
	}
}

func TestBuildDirectoryConcatFilterMapsCoverAfterAudioAndMetadataInputs(t *testing.T) {
	builder := testBuilder()
	first := compatibleMP3(filepath.Join("dir", "01.mp3"), time.Second)
	second := compatibleMP3(filepath.Join("dir", "02.mp3"), time.Second)
	second.NonAudioStreams = 1

	result, err := builder.Build(context.Background(), BuildRequest{
		Input:      audio.Input{Files: []audio.File{first, second}},
		CoverPath:  "cover.jpg",
		OutputPath: "book.m4b",
		DryRun:     true,
	})
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	if !containsInOrder(result.Command,
		"-i", first.Path,
		"-i", second.Path,
		"-i",
		"-i", "cover.jpg",
		"-map", "[bookbind_audio]",
		"-map", "3:v:0",
	) {
		t.Fatalf("command maps fallback inputs incorrectly: %#v", result.Command)
	}
	if !containsInOrder(result.Command, "-map_metadata", "2", "-map_chapters", "2") {
		t.Fatalf("command does not map fallback metadata input: %#v", result.Command)
	}
}

func TestConcatDemuxerCompatibility(t *testing.T) {
	compatible := func() []audio.File {
		return []audio.File{
			compatibleMP3("01.mp3", time.Second),
			compatibleMP3("02.mp3", time.Second),
		}
	}

	tests := []struct {
		name   string
		change func([]audio.File)
		want   bool
	}{
		{name: "same stream parameters", want: true},
		{
			name: "different bitrate is compatible",
			change: func(files []audio.File) {
				files[1].Bitrate = 192000
			},
			want: true,
		},
		{
			name: "different codec",
			change: func(files []audio.File) {
				files[1].Codec = "aac"
			},
		},
		{
			name: "different sample rate",
			change: func(files []audio.File) {
				files[1].SampleRate = 48000
			},
		},
		{
			name: "different sample format",
			change: func(files []audio.File) {
				files[1].SampleFormat = "s16p"
			},
		},
		{
			name: "different channels",
			change: func(files []audio.File) {
				files[1].Channels = 1
				files[1].ChannelLayout = "mono"
			},
		},
		{
			name: "different channel layout",
			change: func(files []audio.File) {
				files[1].ChannelLayout = "2 channels"
			},
		},
		{
			name: "different time base",
			change: func(files []audio.File) {
				files[1].TimeBase = "1/48000"
			},
		},
		{
			name: "embedded cover stream",
			change: func(files []audio.File) {
				files[1].NonAudioStreams = 1
			},
		},
		{
			name: "multiple audio streams",
			change: func(files []audio.File) {
				files[1].AudioStreams = 2
			},
		},
		{
			name: "missing probe data",
			change: func(files []audio.File) {
				files[1].SampleFormat = ""
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			files := compatible()
			if tt.change != nil {
				tt.change(files)
			}
			if got := concatDemuxerCompatible(files); got != tt.want {
				t.Fatalf("concatDemuxerCompatible() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBuildDirectoryWithCoverMapsAttachedPicture(t *testing.T) {
	builder := testBuilder()

	result, err := builder.Build(context.Background(), BuildRequest{
		Input: audio.Input{
			Files: []audio.File{
				compatibleMP3(filepath.Join("dir", "01.mp3"), time.Second),
				compatibleMP3(filepath.Join("dir", "02.mp3"), time.Second),
			},
		},
		CoverPath:  "cover.png",
		OutputPath: "book.m4b",
		DryRun:     true,
	})
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	if !containsInOrder(result.Command, "-i", "cover.png", "-map", "2:v:0") {
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

func compatibleMP3(path string, duration time.Duration) audio.File {
	return audio.File{
		Path:            path,
		Name:            filepath.Base(path),
		Duration:        duration,
		Codec:           "mp3",
		Bitrate:         128000,
		SampleRate:      44100,
		SampleFormat:    "fltp",
		Channels:        2,
		ChannelLayout:   "stereo",
		TimeBase:        "1/14112000",
		AudioStreams:    1,
		NonAudioStreams: 0,
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
