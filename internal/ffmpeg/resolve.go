package ffmpeg

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

func ResolveBinary(name string) string {
	if name == "" {
		return name
	}
	if filepath.IsAbs(name) || filepath.Base(name) != name {
		return name
	}
	if path, err := exec.LookPath(name); err == nil {
		return path
	}
	for _, dir := range candidateDirs() {
		path := filepath.Join(dir, name)
		if isExecutable(path) {
			return path
		}
	}
	return name
}

func candidateDirs() []string {
	dirs := []string{
		"/opt/homebrew/bin",
		"/usr/local/bin",
		"/usr/bin",
		"/opt/local/bin",
	}
	if runtime.GOOS == "linux" {
		dirs = append(dirs, "/snap/bin")
	}
	if exe, err := os.Executable(); err == nil {
		base := filepath.Dir(exe)
		dirs = append(dirs,
			filepath.Join(base, "bin"),
			filepath.Join(base, "..", "Resources", "bin"),
			filepath.Join(base, "..", "Frameworks", "bin"),
		)
	}
	return dirs
}

func isExecutable(path string) bool {
	info, err := os.Stat(path)
	if err != nil || info.IsDir() {
		return false
	}
	if runtime.GOOS == "windows" {
		return true
	}
	return info.Mode()&0o111 != 0
}
