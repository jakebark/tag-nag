package terraform

import (
	"errors"
	"io/fs"
	"path/filepath"
)

// scan looks for tf files
func scan(directoryPath string) (bool, error) {
	found := false
	targetExt := ".tf"

	walkErr := filepath.WalkDir(directoryPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		if !d.IsDir() {
			if filepath.Ext(path) == targetExt {
				found = true
				return fs.ErrNotExist // stop scan immediately
			}
		}
		return nil
	})

	if walkErr != nil && !errors.Is(walkErr, fs.ErrNotExist) {
		return false, walkErr
	}

	return found, nil
}
