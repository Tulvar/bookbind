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

func TestProbeResultIncludesConcatCompatibilityData(t *testing.T) {
	output := ffprobeOutput{
		Streams: []ffprobeStream{
			{
				CodecType:     "audio",
				CodecName:     "mp3",
				Duration:      "3.5",
				BitRate:       "128000",
				SampleRate:    "44100",
				SampleFormat:  "fltp",
				Channels:      2,
				ChannelLayout: "stereo",
				TimeBase:      "1/14112000",
			},
			{CodecType: "video", CodecName: "mjpeg"},
		},
	}

	got := output.probeResult()
	if got.Codec != "mp3" || got.SampleRate != 44100 || got.SampleFormat != "fltp" {
		t.Fatalf("audio format = %#v", got)
	}
	if got.Channels != 2 || got.ChannelLayout != "stereo" || got.TimeBase != "1/14112000" {
		t.Fatalf("audio layout/time base = %#v", got)
	}
	if got.AudioStreams != 1 || got.NonAudioStreams != 1 {
		t.Fatalf("stream counts = audio %d, non-audio %d", got.AudioStreams, got.NonAudioStreams)
	}
	if got.Duration != 3500*time.Millisecond || got.Bitrate != 128000 {
		t.Fatalf("duration/bitrate = %v/%d", got.Duration, got.Bitrate)
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
