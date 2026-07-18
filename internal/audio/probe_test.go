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

func TestFFProbeArgsCountPackets(t *testing.T) {
	args := ffprobeArgs("book.mp3")
	found := false
	for _, arg := range args {
		if arg == "-count_packets" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("ffprobe args do not count packets: %#v", args)
	}
	if got := args[len(args)-1]; got != "book.mp3" {
		t.Fatalf("last ffprobe arg = %q, want input path", got)
	}
}

func TestAccurateMP3Duration(t *testing.T) {
	tests := []struct {
		name     string
		reported time.Duration
		stream   ffprobeStream
		want     time.Duration
	}{
		{
			name:     "replaces bitrate estimate with packet duration",
			reported: 1203417875 * time.Microsecond,
			stream: ffprobeStream{
				CodecName:   "mp3",
				SampleRate:  "44100",
				ReadPackets: "46174",
			},
			want: sampleDuration(46174*1152, 44100),
		},
		{
			name:     "preserves nearby gapless duration",
			reported: 10 * time.Second,
			stream: ffprobeStream{
				CodecName:   "mp3",
				SampleRate:  "44100",
				ReadPackets: "384",
			},
			want: 10 * time.Second,
		},
		{
			name: "uses packets when reported duration is missing",
			stream: ffprobeStream{
				CodecName:   "mp3",
				SampleRate:  "22050",
				ReadPackets: "100",
			},
			want: sampleDuration(100*576, 22050),
		},
		{
			name:     "keeps non-mp3 duration",
			reported: 5 * time.Second,
			stream: ffprobeStream{
				CodecName:   "aac",
				SampleRate:  "44100",
				ReadPackets: "1000",
			},
			want: 5 * time.Second,
		},
		{
			name:     "keeps duration without packet count",
			reported: 5 * time.Second,
			stream: ffprobeStream{
				CodecName:  "mp3",
				SampleRate: "44100",
			},
			want: 5 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := accurateMP3Duration(tt.reported, tt.stream); got != tt.want {
				t.Fatalf("accurateMP3Duration() = %v, want %v", got, tt.want)
			}
		})
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
	if got.HasAttachedPicture {
		t.Fatal("HasAttachedPicture = true for video stream without attached_pic disposition")
	}
	if got.Duration != 3500*time.Millisecond || got.Bitrate != 128000 {
		t.Fatalf("duration/bitrate = %v/%d", got.Duration, got.Bitrate)
	}
}

func TestProbeResultDetectsAttachedPicture(t *testing.T) {
	output := ffprobeOutput{
		Streams: []ffprobeStream{
			{CodecType: "audio", CodecName: "mp3"},
			{CodecType: "video", CodecName: "png"},
			{
				CodecType:   "video",
				CodecName:   "mjpeg",
				Disposition: ffprobeDisposition{AttachedPic: 1},
			},
		},
	}

	got := output.probeResult()
	if !got.HasAttachedPicture {
		t.Fatal("HasAttachedPicture = false, want true")
	}
	if got.AttachedPictureStream != 1 {
		t.Fatalf("AttachedPictureStream = %d, want 1", got.AttachedPictureStream)
	}
}

func TestProbeResultUsesCountedMP3Duration(t *testing.T) {
	output := ffprobeOutput{
		Streams: []ffprobeStream{
			{
				CodecType:   "audio",
				CodecName:   "mp3",
				SampleRate:  "44100",
				ReadPackets: "46174",
			},
		},
		Format: ffprobeFormat{Duration: "1203.417875"},
	}

	got := output.probeResult()
	want := sampleDuration(46174*1152, 44100)
	if got.Duration != want {
		t.Fatalf("Duration = %v, want counted duration %v", got.Duration, want)
	}
}

func TestMP3SamplesPerPacket(t *testing.T) {
	tests := map[int]int64{
		0:     0,
		8000:  576,
		24000: 576,
		32000: 1152,
		44100: 1152,
		48000: 1152,
	}
	for sampleRate, want := range tests {
		if got := mp3SamplesPerPacket(sampleRate); got != want {
			t.Fatalf("mp3SamplesPerPacket(%d) = %d, want %d", sampleRate, got, want)
		}
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
