package app

import (
	"encoding/hex"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type CacheEntry struct {
	Path  string
	Size  int64
	IsDir bool
}

type CacheListResult struct {
	Path    string
	Entries []CacheEntry
	Size    int64
}

type CacheCleanResult struct {
	Path        string
	Removed     int
	RemovedSize int64
}

func CacheDir() (string, error) {
	root, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, "bookbind"), nil
}

func (a *App) ListCache(path string) (CacheListResult, error) {
	cachePath, err := normalizeCachePath(path)
	if err != nil {
		return CacheListResult{}, err
	}
	result := CacheListResult{Path: cachePath, Entries: []CacheEntry{}}
	root, err := os.OpenRoot(cachePath)
	if os.IsNotExist(err) {
		return result, nil
	}
	if err != nil {
		return result, err
	}
	defer root.Close()

	entries, err := cacheFiles(root)
	if err != nil {
		return result, err
	}
	for _, entry := range entries {
		entry.Path = filepath.Join(cachePath, entry.Path)
		result.Entries = append(result.Entries, entry)
		result.Size += entry.Size
	}
	return result, nil
}

func (a *App) CleanCache(path string) (CacheCleanResult, error) {
	cachePath, err := normalizeCachePath(path)
	if err != nil {
		return CacheCleanResult{}, err
	}
	result := CacheCleanResult{Path: cachePath}
	root, err := os.OpenRoot(cachePath)
	if os.IsNotExist(err) {
		return result, nil
	}
	if err != nil {
		return result, err
	}
	defer root.Close()

	entries, err := cacheFiles(root)
	if err != nil {
		return result, err
	}
	for _, entry := range entries {
		if err := root.Remove(entry.Path); os.IsNotExist(err) {
			continue
		} else if err != nil {
			return result, fmt.Errorf("remove cached file %s: %w", entry.Path, err)
		}
		result.Removed++
		result.RemovedSize += entry.Size
	}
	return result, nil
}

func normalizeCachePath(path string) (string, error) {
	if path != "" {
		return filepath.Abs(path)
	}
	return CacheDir()
}

func cacheFiles(root *os.Root) ([]CacheEntry, error) {
	var files []CacheEntry
	for _, dir := range []string{"covers", "providers"} {
		info, err := root.Lstat(dir)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("inspect cache directory %s: %w", dir, err)
		}
		if !info.IsDir() {
			continue
		}
		entries, err := fs.ReadDir(root.FS(), dir)
		if err != nil {
			return nil, fmt.Errorf("list cache directory %s: %w", dir, err)
		}
		for _, entry := range entries {
			if !entry.Type().IsRegular() || !isCacheFileName(dir, entry.Name()) {
				continue
			}
			info, err := entry.Info()
			if os.IsNotExist(err) {
				continue
			}
			if err != nil {
				return nil, fmt.Errorf("inspect cached file %s: %w", entry.Name(), err)
			}
			if info.Mode().IsRegular() {
				files = append(files, CacheEntry{Path: filepath.Join(dir, entry.Name()), Size: info.Size()})
			}
		}
	}
	return files, nil
}

func isCacheFileName(dir, name string) bool {
	ext := filepath.Ext(name)
	stem := strings.TrimSuffix(name, ext)
	if dir == "providers" {
		provider, key, ok := strings.Cut(stem, "-")
		if !ok || ext != ".json" {
			return false
		}
		if canonical, err := CanonicalProviderName(provider); err != nil || canonical != provider {
			return false
		}
		op, digest, ok := strings.Cut(key, "-")
		return ok && (op == "get" || op == "search") && isCacheDigest(digest)
	}
	return (ext == ".jpg" || ext == ".jpeg" || ext == ".png") && isCacheDigest(stem)
}

func isCacheDigest(value string) bool {
	if len(value) != 64 || value != strings.ToLower(value) {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
}
