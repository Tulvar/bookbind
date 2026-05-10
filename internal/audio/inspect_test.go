package audio

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestInspectDirectoryFindsAndSortsMP3Files(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, filepath.Join(dir, "02.mp3"))
	writeTestFile(t, filepath.Join(dir, "01.MP3"))
	writeTestFile(t, filepath.Join(dir, "cover.jpg"))

	input, err := NewInspector(WithProber(fakeProber{})).Inspect(context.Background(), dir)
	if err != nil {
		t.Fatalf("Inspect() error = %v", err)
	}

	if got, want := len(input.Files), 2; got != want {
		t.Fatalf("len(files) = %d, want %d", got, want)
	}
	if got, want := input.Files[0].Name, "01.MP3"; got != want {
		t.Fatalf("first file = %q, want %q", got, want)
	}
	if got, want := input.Files[1].Name, "02.mp3"; got != want {
		t.Fatalf("second file = %q, want %q", got, want)
	}
}

func TestInspectRejectsNonMP3File(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "book.wav")
	writeTestFile(t, path)

	_, err := NewInspector(WithProber(fakeProber{})).Inspect(context.Background(), path)
	if err == nil {
		t.Fatal("Inspect() error = nil, want error")
	}
}

func TestInspectFileAddsProbeData(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "book.mp3")
	writeTestFile(t, path)

	input, err := NewInspector(WithProber(fakeProber{})).Inspect(context.Background(), path)
	if err != nil {
		t.Fatalf("Inspect() error = %v", err)
	}

	file := input.Files[0]
	if got, want := file.Duration, 3*time.Second; got != want {
		t.Fatalf("Duration = %v, want %v", got, want)
	}
	if got, want := file.Codec, "mp3"; got != want {
		t.Fatalf("Codec = %q, want %q", got, want)
	}
}

func writeTestFile(t *testing.T, path string) {
	t.Helper()
	if err := os.WriteFile(path, []byte("test"), 0o644); err != nil {
		t.Fatalf("write test file: %v", err)
	}
}

type fakeProber struct{}

func (fakeProber) Probe(context.Context, string) (ProbeResult, error) {
	return ProbeResult{
		Duration: 3 * time.Second,
		Codec:    "mp3",
		Bitrate:  128000,
		Channels: 2,
		Tags: EmbeddedTags{
			Title:  "Embedded Title",
			Artist: "Embedded Artist",
		},
	}, nil
}
