package terraform

import (
	"errors"
	"io/fs"
	"path/filepath"
)

// CheckForTerraformFiles checks if any .tf files exist in the directory
func scan(dirPath string) (bool, error) {
	found := false
	targetExt := ".tf"

	walkErr := filepath.WalkDir(dirPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		if !d.IsDir() {
			if filepath.Ext(path) == targetExt {
				found = true
				return fs.ErrNotExist
			}
		}
		return nil
	})

	if walkErr != nil && !errors.Is(walkErr, fs.ErrNotExist) {
		return false, walkErr
	}

	return found, nil
}
