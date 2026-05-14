package audio

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func TestParseDuration(t *testing.T) {
	got := parseDuration("12.500000")
	want := 12*time.Second + 500*time.Millisecond
	if got != want {
		t.Fatalf("parseDuration() = %v, want %v", got, want)
	}
}

func TestFirstAudioStream(t *testing.T) {
	output := ffprobeOutput{
		Streams: []ffprobeStream{
			{CodecType: "video", CodecName: "mjpeg"},
			{CodecType: "audio", CodecName: "mp3", Channels: 2},
		},
	}

	stream := output.firstAudioStream()
	if stream.CodecName != "mp3" {
		t.Fatalf("CodecName = %q, want mp3", stream.CodecName)
	}
}

func TestEmbeddedTagsNormalizesKeys(t *testing.T) {
	got := embeddedTags(ffTags{
		"TITLE":        " Night Watch ",
		"Artist":       "Sergey Lukyanenko",
		"album artist": "Album Artist",
		"YEAR":         "1998",
	})

	if got.Title != "Night Watch" {
		t.Fatalf("Title = %q", got.Title)
	}
	if got.Artist != "Sergey Lukyanenko" {
		t.Fatalf("Artist = %q", got.Artist)
	}
	if got.AlbumArtist != "Album Artist" {
		t.Fatalf("AlbumArtist = %q", got.AlbumArtist)
	}
	if got.Date != "1998" {
		t.Fatalf("Date = %q", got.Date)
	}
}

func TestChaptersParsesFFProbeChapters(t *testing.T) {
	got := chapters([]ffprobeChapter{
		{StartTime: "0.000000", EndTime: "60.000000", Tags: ffTags{"title": "Intro"}},
		{StartTime: "60.000000", EndTime: "120.000000"},
	})

	if len(got) != 2 {
		t.Fatalf("len = %d, want 2", len(got))
	}
	if got[0].Title != "Intro" || got[0].Start != 0 || got[0].End != time.Minute {
		t.Fatalf("first chapter = %#v", got[0])
	}
	if got[1].Title != "Chapter 002" || got[1].Start != time.Minute || got[1].End != 2*time.Minute {
		t.Fatalf("second chapter = %#v", got[1])
	}
}

func TestNewFFProbeResolvesFromPath(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, testExecutableName("ffprobe"))
	if err := os.WriteFile(path, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatalf("write ffprobe: %v", err)
	}
	t.Setenv("PATH", dir)

	prober := NewFFProbe("ffprobe")
	if prober.Path != path {
		t.Fatalf("Path = %q, want %q", prober.Path, path)
	}
}

func testExecutableName(name string) string {
	if runtime.GOOS == "windows" {
		return name + ".exe"
	}
	return name
}
