//go:build integration

package main

import (
	"bytes"
	"context"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestConvertRelativeDirectory(t *testing.T) {
	ffmpeg, err := exec.LookPath("ffmpeg")
	if err != nil {
		t.Skip("ffmpeg is required")
	}
	ffprobe, err := exec.LookPath("ffprobe")
	if err != nil {
		t.Skip("ffprobe is required")
	}

	t.Chdir(t.TempDir())
	const input = "tracks with spaces"
	if err := os.Mkdir(input, 0o700); err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	track := filepath.Join(input, "01 - opening's chapter.mp3")
	output, err := exec.CommandContext(ctx, ffmpeg, "-v", "error", "-f", "lavfi", "-i",
		"sine=duration=0.2", "-c:a", "libmp3lame", track).CombinedOutput()
	if err != nil {
		if strings.Contains(string(output), "Unknown encoder") {
			t.Skip("ffmpeg with libmp3lame is required to generate the test MP3")
		}
		t.Fatalf("create test MP3: %v\n%s", err, output)
	}
	data, err := os.ReadFile(track)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(input, "02.mp3"), data, 0o600); err != nil {
		t.Fatal(err)
	}

	var log bytes.Buffer
	if err := run(ctx, []string{"convert", input, "--output", "book.m4b"}, &log, io.Discard); err != nil {
		t.Fatalf("convert relative directory: %v", err)
	}
	if !strings.Contains(log.String(), "-f concat") {
		t.Fatalf("expected concat demuxer, got %s", &log)
	}
	output, err = exec.CommandContext(ctx, ffprobe, "-v", "error", "-show_entries",
		"chapter=id", "-of", "csv=p=0", "book.m4b").CombinedOutput()
	if err != nil || len(strings.Fields(string(output))) != 2 {
		t.Fatalf("expected two output chapters, got %s (%v)", output, err)
	}
}
