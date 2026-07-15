//go:build !windows

package m4b

import "os"

func renameNoReplace(oldPath, newPath string) error {
	if err := os.Link(oldPath, newPath); err != nil {
		return err
	}
	_ = os.Remove(oldPath)
	return nil
}
