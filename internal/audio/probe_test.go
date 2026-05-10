package audio

import (
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
