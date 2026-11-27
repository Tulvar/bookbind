package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCacheOnlyManagesBookbindFiles(t *testing.T) {
	dir := t.TempDir()
	digest := strings.Repeat("a", 64)
	files := map[string]string{
		"library/book.mp3":                              "audio",
		"providers/notes.json":                          "notes",
		"providers/openlibrary-get-" + digest + ".json": "123",
		"covers/" + digest + ".png":                     "45",
		"covers/manual.jpg":                             "manual",
		"covers/nested/" + digest + ".jpg":              "nested",
	}
	for name, data := range files {
		path := filepath.Join(dir, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(data), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	application := New()
	list, err := application.ListCache(dir)
	if err != nil {
		t.Fatal(err)
	}
	if list.Path != dir || len(list.Entries) != 2 || list.Size != 5 {
		t.Fatalf("ListCache() = %+v, want two files totaling 5 bytes", list)
	}

	clean, err := application.CleanCache(dir)
	if err != nil {
		t.Fatal(err)
	}
	if clean.Removed != 2 || clean.RemovedSize != 5 {
		t.Fatalf("CleanCache() = %+v, want two files totaling 5 bytes", clean)
	}
	for name, want := range files {
		path := filepath.Join(dir, filepath.FromSlash(name))
		data, err := os.ReadFile(path)
		if want == "123" || want == "45" {
			if !os.IsNotExist(err) {
				t.Errorf("expected %s to be removed, got %v", name, err)
			}
		} else if err != nil || string(data) != want {
			t.Errorf("expected %s to be preserved, got %q, %v", name, data, err)
		}
	}
}

func TestCleanCacheSkipsSymlinks(t *testing.T) {
	dir, outside := t.TempDir(), t.TempDir()
	name := strings.Repeat("a", 64) + ".jpg"
	target := filepath.Join(outside, name)
	if err := os.WriteFile(target, []byte("keep"), 0o600); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(dir, "covers")
	if err := os.Symlink(outside, link); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	result, err := New().CleanCache(dir)
	if err != nil || result.Removed != 0 {
		t.Fatalf("CleanCache() = %+v, %v", result, err)
	}
	if data, err := os.ReadFile(target); err != nil || string(data) != "keep" {
		t.Fatalf("linked file was modified: %q, %v", data, err)
	}
	if info, err := os.Lstat(link); err != nil || info.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("symlink was modified: %v, %v", info, err)
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
