package app

import (
	"os"
	"path/filepath"
	"testing"
)

func TestListCacheReturnsEntriesAndSize(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "one.cache"), []byte("123"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	nested := filepath.Join(dir, "providers")
	if err := os.Mkdir(nested, 0o755); err != nil {
		t.Fatalf("Mkdir() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(nested, "two.cache"), []byte("45"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	result, err := New().ListCache(dir)
	if err != nil {
		t.Fatalf("ListCache() error = %v", err)
	}

	if result.Path != dir {
		t.Fatalf("Path = %q", result.Path)
	}
	if len(result.Entries) != 2 {
		t.Fatalf("entries len = %d", len(result.Entries))
	}
	if result.Size != 5 {
		t.Fatalf("Size = %d", result.Size)
	}
}

func TestListCacheAllowsMissingDirectory(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing")

	result, err := New().ListCache(path)
	if err != nil {
		t.Fatalf("ListCache() error = %v", err)
	}

	if result.Path != path {
		t.Fatalf("Path = %q", result.Path)
	}
	if len(result.Entries) != 0 {
		t.Fatalf("entries len = %d", len(result.Entries))
	}
}

func TestCleanCacheRemovesEntries(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "one.cache")
	if err := os.WriteFile(filePath, []byte("123"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	result, err := New().CleanCache(dir)
	if err != nil {
		t.Fatalf("CleanCache() error = %v", err)
	}

	if result.Removed != 1 {
		t.Fatalf("Removed = %d", result.Removed)
	}
	if result.RemovedSize != 3 {
		t.Fatalf("RemovedSize = %d", result.RemovedSize)
	}
	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		t.Fatalf("cache file still exists or unexpected stat error: %v", err)
	}
	if _, err := os.Stat(dir); err != nil {
		t.Fatalf("cache dir should remain: %v", err)
	}
}
