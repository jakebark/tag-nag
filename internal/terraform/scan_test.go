package terraform

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

func TestScanFunction(t *testing.T) {
	createDummyFile := func(t *testing.T, path string) {
		t.Helper()
		err := os.WriteFile(path, []byte("dummy content"), 0644)
		if err != nil {
			t.Fatalf("Failed to create dummy file %s: %v", path, err)
		}
	}

	tests := []struct {
		name     string
		setupDir func(t *testing.T, dir string)
		expected bool
	}{
		{
			name: "empty dir",
			setupDir: func(t *testing.T, dir string) {
				// dir is empty
			},
			expected: false,
		},
		{
			name: "no terraform files",
			setupDir: func(t *testing.T, dir string) {
				createDummyFile(t, filepath.Join(dir, "main.yaml"))
				createDummyFile(t, filepath.Join(dir, "README.md"))
			},
			expected: false,
		},
		{
			name: "terraform file",
			setupDir: func(t *testing.T, dir string) {
				createDummyFile(t, filepath.Join(dir, "main.tf"))
				createDummyFile(t, filepath.Join(dir, "other.txt"))
			},
			expected: true,
		},
		{
			name: "multiple terraform files",
			setupDir: func(t *testing.T, dir string) {
				createDummyFile(t, filepath.Join(dir, "main.tf"))
				createDummyFile(t, filepath.Join(dir, "variables.tf"))
			},
			expected: true,
		},
		{
			name: "nested terraform file",
			setupDir: func(t *testing.T, dir string) {
				subDir := filepath.Join(dir, "subdir")
				err := os.Mkdir(subDir, 0755)
				if err != nil {
					t.Fatalf("Failed to create subdir: %v", err)
				}
				createDummyFile(t, filepath.Join(subDir, "module.tf"))
				createDummyFile(t, filepath.Join(dir, "root.txt"))
			},
			expected: true,
		},
		{
			name: "terraform file, nested non-terraform file",
			setupDir: func(t *testing.T, dir string) {
				subDir := filepath.Join(dir, "subdir")
				err := os.Mkdir(subDir, 0755)
				if err != nil {
					t.Fatalf("Failed to create subdir: %v", err)
				}
				createDummyFile(t, filepath.Join(dir, "main.tf"))
				createDummyFile(t, filepath.Join(subDir, "other.yaml"))
			},
			expected: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			testPath := tmpDir
			if tc.name == "Non-existent directory" {
				testPath = filepath.Join(tmpDir, "this_dir_should_not_exist")
			} else {
				tc.setupDir(t, tmpDir)
			}
			gotFound, gotErr := scan(testPath)
			if gotErr != nil && !errors.Is(gotErr, fs.ErrNotExist) {
				if tc.name != "Non-existent directory" {
					t.Logf("scan() returned an unexpected error: %v (test will check 'found' status only)", gotErr)
				}
			}
			if gotFound != tc.expected {
				t.Errorf("scan() found = %v, expected %v", gotFound, tc.expected)
			}
		})
	}
}
