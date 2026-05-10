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
