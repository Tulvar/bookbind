package main

import (
	"path/filepath"
	"testing"
	"time"
)

func TestBundledToolPaths(t *testing.T) {
	root := t.TempDir()
	for _, test := range []struct {
		goos      string
		exe       string
		dir       string
		extension string
	}{
		{"darwin", "bookbind.app/Contents/MacOS/bookbind", "bookbind.app/Contents/Resources/bin", ""},
		{"linux", "bookbind", "bin", ""},
		{"windows", "bookbind.exe", "bin", ".exe"},
	} {
		t.Run(test.goos, func(t *testing.T) {
			ffmpeg, ffprobe := bundledToolPaths(filepath.Join(root, filepath.FromSlash(test.exe)), test.goos)
			dir := filepath.Join(root, filepath.FromSlash(test.dir))
			if ffmpeg != filepath.Join(dir, "ffmpeg"+test.extension) || ffprobe != filepath.Join(dir, "ffprobe"+test.extension) {
				t.Fatalf("bundledToolPaths() = %q, %q, want both tools in %q", ffmpeg, ffprobe, dir)
			}
		})
	}
}

func TestFormatDuration(t *testing.T) {
	tests := map[time.Duration]string{
		0:                           "",
		65 * time.Second:            "1m 05s",
		2*time.Hour + 3*time.Second: "2h 00m 03s",
	}

	for duration, want := range tests {
		if got := formatDuration(duration); got != want {
			t.Fatalf("formatDuration(%s) = %q, want %q", duration, got, want)
		}
	}
}

func TestFormatBitrate(t *testing.T) {
	if got := formatBitrate(128000); got != "128 kbps" {
		t.Fatalf("formatBitrate() = %q", got)
	}
}
