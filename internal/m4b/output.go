package m4b

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func createTemporaryOutput(outputPath string) (string, error) {
	dir := filepath.Dir(outputPath)
	base := strings.TrimSuffix(filepath.Base(outputPath), filepath.Ext(outputPath))
	file, err := os.CreateTemp(dir, "."+base+".bookbind-*.m4b")
	if err != nil {
		return "", fmt.Errorf("create temporary output next to %s: %w", outputPath, err)
	}
	path := file.Name()
	if err := file.Close(); err != nil {
		_ = os.Remove(path)
		return "", fmt.Errorf("close temporary output %s: %w", path, err)
	}
	return path, nil
}

func publishTemporaryOutput(temporaryPath, outputPath string, overwrite bool) error {
	info, err := os.Stat(temporaryPath)
	if err != nil {
		return fmt.Errorf("inspect temporary output %s: %w", temporaryPath, err)
	}
	if !info.Mode().IsRegular() || info.Size() == 0 {
		return fmt.Errorf("ffmpeg produced an empty or invalid temporary output: %s", temporaryPath)
	}

	mode := os.FileMode(0o644)
	if overwrite {
		if current, statErr := os.Stat(outputPath); statErr == nil && current.Mode().IsRegular() {
			mode = current.Mode().Perm()
		}
	}
	if err := os.Chmod(temporaryPath, mode); err != nil {
		return fmt.Errorf("set temporary output permissions %s: %w", temporaryPath, err)
	}
	if err := syncFile(temporaryPath); err != nil {
		return err
	}

	if overwrite {
		err = os.Rename(temporaryPath, outputPath)
	} else {
		err = renameNoReplace(temporaryPath, outputPath)
	}
	if err != nil {
		return fmt.Errorf("publish output %s: %w", outputPath, err)
	}
	return nil
}

func syncFile(path string) error {
	file, err := os.OpenFile(path, os.O_RDWR, 0)
	if err != nil {
		return fmt.Errorf("open temporary output for sync %s: %w", path, err)
	}
	if err := file.Sync(); err != nil {
		_ = file.Close()
		return fmt.Errorf("sync temporary output %s: %w", path, err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("close temporary output after sync %s: %w", path, err)
	}
	return nil
}
