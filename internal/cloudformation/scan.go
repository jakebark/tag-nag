package cloudformation

import (
	"errors"
	"io/fs"
	"path/filepath"
)

// scan looks for cfn files
func scan(dirPath string) (bool, error) {
	found := false
	targetExts := map[string]bool{
		".yaml": true,
		".yml":  true,
		".json": true,
	}
	walkErr := filepath.WalkDir(dirPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		if !d.IsDir() {
			if targetExts[filepath.Ext(path)] {
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
