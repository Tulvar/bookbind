package app

import (
	"fmt"
	"os"
	"path/filepath"
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

	entries, err := os.ReadDir(cachePath)
	if os.IsNotExist(err) {
		return CacheListResult{Path: cachePath}, nil
	}
	if err != nil {
		return CacheListResult{}, err
	}

	result := CacheListResult{Path: cachePath, Entries: make([]CacheEntry, 0, len(entries))}
	for _, entry := range entries {
		fullPath := filepath.Join(cachePath, entry.Name())
		size, err := cacheEntrySize(fullPath)
		if err != nil {
			return CacheListResult{}, err
		}
		result.Entries = append(result.Entries, CacheEntry{
			Path:  fullPath,
			Size:  size,
			IsDir: entry.IsDir(),
		})
		result.Size += size
	}
	return result, nil
}

func (a *App) CleanCache(path string) (CacheCleanResult, error) {
	list, err := a.ListCache(path)
	if err != nil {
		return CacheCleanResult{}, err
	}

	result := CacheCleanResult{Path: list.Path, RemovedSize: list.Size}
	for _, entry := range list.Entries {
		if err := os.RemoveAll(entry.Path); err != nil {
			return CacheCleanResult{}, err
		}
		result.Removed++
	}
	return result, nil
}

func normalizeCachePath(path string) (string, error) {
	if path != "" {
		return path, nil
	}
	return CacheDir()
}

func cacheEntrySize(path string) (int64, error) {
	var size int64
	err := filepath.WalkDir(path, func(_ string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		size += info.Size()
		return nil
	})
	if err != nil {
		return 0, fmt.Errorf("cache size %s: %w", path, err)
	}
	return size, nil
}
