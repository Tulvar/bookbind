package m4b

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/Tulvar/bookbind/internal/audio"
)

func TestBuildDryRunSingleFilePlansCommand(t *testing.T) {
	builder := NewBuilder("ffmpeg")

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
	builder := NewBuilder("ffmpeg")
	builder.Runner = runner

	_, err := builder.Build(context.Background(), BuildRequest{
		Input: audio.Input{
			Files: []audio.File{{Path: "book.mp3"}},
		},
		OutputPath: "book.m4b",
		Overwrite:  true,
	})
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	command := append([]string{runner.name}, runner.args...)
	if !containsInOrder(command, "ffmpeg", "-y", "-i", "book.mp3", "-i") {
		t.Fatalf("command does not contain expected inputs: %#v", command)
	}
	if !containsInOrder(command, "-c:a", "aac", "-b:a", "64k", "book.m4b") {
		t.Fatalf("command does not contain conversion args: %#v", command)
	}
}

func TestBuildSingleFileWithCoverMapsAttachedPicture(t *testing.T) {
	builder := NewBuilder("ffmpeg")

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
	builder := NewBuilder("ffmpeg")

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

func TestBuildSingleFileUsesEmbeddedChapters(t *testing.T) {
	builder := NewBuilder("ffmpeg")

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
	builder := NewBuilder("ffmpeg")

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
	builder := NewBuilder("ffmpeg")

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

	if got, want := result.Command[0], "ffmpeg"; got != want {
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
	builder := NewBuilder("ffmpeg")

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
	builder := NewBuilder("ffmpeg")

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
	return nil
}
