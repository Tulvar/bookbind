//go:build windows

package m4b

import "syscall"

func renameNoReplace(oldPath, newPath string) error {
	return syscall.Rename(oldPath, newPath)
}
