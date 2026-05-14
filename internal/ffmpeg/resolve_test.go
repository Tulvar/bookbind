package ffmpeg

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestResolveBinaryKeepsExplicitPath(t *testing.T) {
	path := filepath.Join("custom", "ffmpeg")
	if got := ResolveBinary(path); got != path {
		t.Fatalf("ResolveBinary() = %q, want %q", got, path)
	}
}

func TestResolveBinaryFallsBackToName(t *testing.T) {
	t.Setenv("PATH", "")
	if got := ResolveBinary("definitely-not-bookbind-binary"); got != "definitely-not-bookbind-binary" {
		t.Fatalf("ResolveBinary() = %q, want fallback name", got)
	}
}

func TestResolveBinaryUsesPath(t *testing.T) {
	dir := t.TempDir()
	name := testBinaryName("bookbind-test-tool")
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatalf("write tool: %v", err)
	}
	t.Setenv("PATH", dir)

	if got := ResolveBinary("bookbind-test-tool"); got != path {
		t.Fatalf("ResolveBinary() = %q, want %q", got, path)
	}
}

func testBinaryName(name string) string {
	if runtime.GOOS == "windows" {
		return name + ".exe"
	}
	return name
}
