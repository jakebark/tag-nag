package inputs

import (
	"os"
	"reflect"
	"testing"

	"github.com/jakebark/tag-nag/internal/shared"
)

func TestProcessConfigFile(t *testing.T) {
	testCases := []struct {
		name              string
		configFile        string
		expectedError     bool
		expectedTags      int
		expectedOwner     bool
		expectedEnvValues []string
		expectedSettings  Settings
		expectedSkips     []string
	}{
		{
			name:          "valid basic config",
			configFile:    "../../testdata/config/valid-basic.yml",
			expectedError: false,
			expectedTags:  2,
			expectedOwner: true,
			expectedSettings: Settings{
				CaseInsensitive: false,
				DryRun:          false,
				CfnSpec:         "",
			},
			expectedSkips: []string{},
		},
		{
			name:              "valid with environment values",
			configFile:        "../../testdata/config/valid-with-values.yml",
			expectedError:     false,
			expectedTags:      3,
			expectedOwner:     true,
			expectedEnvValues: []string{"Dev", "Test", "Prod"},
			expectedSettings: Settings{
				CaseInsensitive: false,
				DryRun:          false,
				CfnSpec:         "",
			},
			expectedSkips: []string{},
		},
		{
			name:          "complete config with all fields",
			configFile:    "../../testdata/config/valid-complete.yml",
			expectedError: false,
			expectedTags:  3,
			expectedOwner: true,
			expectedSettings: Settings{
				CaseInsensitive: true,
				DryRun:          false,
				CfnSpec:         "/path/to/spec.json",
			},
			expectedSkips: []string{"*.tmp", ".terraform", "test-data/**"},
		},
		{
			name:          "minimal config",
			configFile:    "../../testdata/config/valid-minimal.yml",
			expectedError: false,
			expectedTags:  1,
			expectedOwner: true,
			expectedSettings: Settings{
				CaseInsensitive: false,
				DryRun:          false,
				CfnSpec:         "",
			},
			expectedSkips: []string{},
		},
		{
			name:          "empty settings and skip sections",
			configFile:    "../../testdata/config/valid-empty-settings.yml",
			expectedError: false,
			expectedTags:  1,
			expectedOwner: true,
			expectedSettings: Settings{
				CaseInsensitive: false,
				DryRun:          false,
				CfnSpec:         "",
			},
			expectedSkips: []string{},
		},
		{
			name:          "yaml extension works",
			configFile:    "../../testdata/config/valid-alt-extension.yaml",
			expectedError: false,
			expectedTags:  1,
			expectedOwner: true,
			expectedSettings: Settings{
				CaseInsensitive: false,
				DryRun:          false,
				CfnSpec:         "",
			},
			expectedSkips: []string{},
		},
		{
			name:          "empty file",
			configFile:    "../../testdata/config/invalid-empty.yml",
			expectedError: false,
			expectedTags:  0,
			expectedOwner: false,
			expectedSettings: Settings{
				CaseInsensitive: false,
				DryRun:          false,
				CfnSpec:         "",
			},
			expectedSkips: []string{},
		},
		{
			name:          "yaml syntax error",
			configFile:    "../../testdata/config/invalid-yaml-syntax.yml",
			expectedError: true,
		},
		{
			name:          "missing tags section",
			configFile:    "../../testdata/config/invalid-no-tags.yml",
			expectedError: false,
			expectedTags:  0,
			expectedOwner: false,
			expectedSettings: Settings{
				CaseInsensitive: false,
				DryRun:          true,
				CfnSpec:         "",
			},
			expectedSkips: []string{"*.tmp"},
		},
		{
			name:          "wrong tags structure",
			configFile:    "../../testdata/config/invalid-wrong-structure.yml",
			expectedError: true,
		},
		{
			name:          "bad values type",
			configFile:    "../../testdata/config/invalid-bad-values.yml",
			expectedError: true,
		},
		{
			name:          "nonexistent file",
			configFile:    "../../testdata/config/does-not-exist.yml",
			expectedError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config, err := processConfigFile(tc.configFile)

			// Check error expectation
			if tc.expectedError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if config == nil {
				t.Errorf("Expected config but got nil")
				return
			}

			// Check number of tags
			if len(config.Tags) != tc.expectedTags {
				t.Errorf("Expected %d tags, got %d", tc.expectedTags, len(config.Tags))
			}

			// Check if Owner tag exists
			hasOwner := false
			var envValues []string
			for _, tag := range config.Tags {
				if tag.Key == "Owner" {
					hasOwner = true
				}
				if tag.Key == "Environment" {
					envValues = tag.Values
				}
			}

			if hasOwner != tc.expectedOwner {
				t.Errorf("Expected Owner tag: %v, got: %v", tc.expectedOwner, hasOwner)
			}

			// Check Environment values if specified
			if tc.expectedEnvValues != nil {
				if !reflect.DeepEqual(envValues, tc.expectedEnvValues) {
					t.Errorf("Expected Environment values %v, got %v", tc.expectedEnvValues, envValues)
				}
			}

			// Check settings
			if config.Settings != tc.expectedSettings {
				t.Errorf("Expected settings %+v, got %+v", tc.expectedSettings, config.Settings)
			}

			// Check skip patterns (handle nil vs empty slice)
			configSkips := config.Skip
			if configSkips == nil {
				configSkips = []string{}
			}
			if !reflect.DeepEqual(configSkips, tc.expectedSkips) {
				t.Errorf("Expected skip patterns %v, got %v", tc.expectedSkips, configSkips)
			}
		})
	}
}

func TestFindAndLoadConfigFile(t *testing.T) {
	// Save current directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir)

	testCases := []struct {
		name           string
		setupFunc      func(t *testing.T, tempDir string)
		expectedError  bool
		expectedConfig bool
		expectedTags   int
	}{
		{
			name: "no config files present",
			setupFunc: func(t *testing.T, tempDir string) {
				// Empty directory
			},
			expectedError:  false,
			expectedConfig: false,
		},
		{
			name: ".tag-nag.yml present",
			setupFunc: func(t *testing.T, tempDir string) {
				content := "tags:\n  - key: Owner\n"
				err := os.WriteFile(".tag-nag.yml", []byte(content), 0644)
				if err != nil {
					t.Fatalf("Failed to create test config: %v", err)
				}
			},
			expectedError:  false,
			expectedConfig: true,
			expectedTags:   1,
		},
		{
			name: ".tag-nag.yaml present",
			setupFunc: func(t *testing.T, tempDir string) {
				content := "tags:\n  - key: Project\n"
				err := os.WriteFile(".tag-nag.yaml", []byte(content), 0644)
				if err != nil {
					t.Fatalf("Failed to create test config: %v", err)
				}
			},
			expectedError:  false,
			expectedConfig: true,
			expectedTags:   1,
		},
		{
			name: "both files present (should prefer .yml)",
			setupFunc: func(t *testing.T, tempDir string) {
				ymlContent := "tags:\n  - key: Owner\n  - key: Project\n"
				yamlContent := "tags:\n  - key: Environment\n"

				err := os.WriteFile(".tag-nag.yml", []byte(ymlContent), 0644)
				if err != nil {
					t.Fatalf("Failed to create .yml config: %v", err)
				}

				err = os.WriteFile(".tag-nag.yaml", []byte(yamlContent), 0644)
				if err != nil {
					t.Fatalf("Failed to create .yaml config: %v", err)
				}
			},
			expectedError:  false,
			expectedConfig: true,
			expectedTags:   2, // Should use .yml file (2 tags), not .yaml file (1 tag)
		},
		{
			name: "invalid config file",
			setupFunc: func(t *testing.T, tempDir string) {
				content := "tags:\n  - key: Owner\n    values: [invalid"
				err := os.WriteFile(".tag-nag.yml", []byte(content), 0644)
				if err != nil {
					t.Fatalf("Failed to create invalid config: %v", err)
				}
			},
			expectedError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create temporary directory for each test
			tempDir, err := os.MkdirTemp("", "tag-nag-test")
			if err != nil {
				t.Fatalf("Failed to create temp dir: %v", err)
			}
			defer os.RemoveAll(tempDir)

			// Change to temp directory
			err = os.Chdir(tempDir)
			if err != nil {
				t.Fatalf("Failed to change to temp dir: %v", err)
			}

			// Setup test scenario
			tc.setupFunc(t, tempDir)

			// Test the function
			config, err := FindAndLoadConfigFile()

			// Check error expectation
			if tc.expectedError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			// Check config presence
			if tc.expectedConfig {
				if config == nil {
					t.Errorf("Expected config but got nil")
					return
				}
				if len(config.Tags) != tc.expectedTags {
					t.Errorf("Expected %d tags, got %d", tc.expectedTags, len(config.Tags))
				}
			} else {
				if config != nil {
					t.Errorf("Expected no config but got: %+v", config)
				}
			}
		})
	}
}

func TestConvertToTagMap(t *testing.T) {
	testCases := []struct {
		name     string
		config   Config
		expected shared.TagMap
	}{
		{
			name: "empty config",
			config: Config{
				Tags: []TagDefinition{},
			},
			expected: shared.TagMap{},
		},
		{
			name: "tags without values",
			config: Config{
				Tags: []TagDefinition{
					{Key: "Owner"},
					{Key: "Project"},
				},
			},
			expected: shared.TagMap{
				"Owner":   []string{},
				"Project": []string{},
			},
		},
		{
			name: "tags with values",
			config: Config{
				Tags: []TagDefinition{
					{Key: "Owner"},
					{Key: "Environment", Values: []string{"Dev", "Test", "Prod"}},
					{Key: "Project"},
				},
			},
			expected: shared.TagMap{
				"Owner":       []string{},
				"Environment": []string{"Dev", "Test", "Prod"},
				"Project":     []string{},
			},
		},
		{
			name: "mixed values",
			config: Config{
				Tags: []TagDefinition{
					{Key: "Owner", Values: []string{"Alice", "Bob"}},
					{Key: "Environment", Values: []string{"Prod"}},
					{Key: "Project"},
				},
			},
			expected: shared.TagMap{
				"Owner":       []string{"Alice", "Bob"},
				"Environment": []string{"Prod"},
				"Project":     []string{},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.config.convertToTagMap()

			// Check each key individually to handle nil vs empty slice differences
			if len(result) != len(tc.expected) {
				t.Errorf("Expected %d keys, got %d", len(tc.expected), len(result))
				return
			}

			for key, expectedValues := range tc.expected {
				actualValues, exists := result[key]
				if !exists {
					t.Errorf("Expected key %s not found in result", key)
					continue
				}

				// Handle nil vs empty slice
				if actualValues == nil {
					actualValues = []string{}
				}
				if expectedValues == nil {
					expectedValues = []string{}
				}

				if !reflect.DeepEqual(actualValues, expectedValues) {
					t.Errorf("Key %s: expected %v, got %v", key, expectedValues, actualValues)
				}
			}
		})
	}
}

