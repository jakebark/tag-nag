package cloudformation

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/jakebark/tag-nag/internal/config"
	"github.com/jakebark/tag-nag/internal/shared"
)

// ProcessDirectory walks all cfn files in a directory, then returns violations
func ProcessDirectory(dirPath string, requiredTags map[string][]string, caseInsensitive bool, specFilePath string, skip []string) []shared.Violation {
	hasFiles, err := scan(dirPath)
	if err != nil {
		return nil
	}
	if !hasFiles {
		return nil
	}

	// log.Println("\nCloudFormation files found")
	var allViolations []shared.Violation

	var taggable map[string]bool
	if specFilePath != "" {
		var loadSpecErr error
		taggable, loadSpecErr = loadTaggableResourcesFromSpec(specFilePath)
		if loadSpecErr != nil {
			log.Printf("Warning: Could not load or parse --cfn-spec file '%s': %v.", specFilePath, loadSpecErr)
			taggable = nil
		} else {
			// log.Println("Parsing CloudFormation spec file.")
		}
	}

	walkErr := filepath.Walk(dirPath, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			log.Printf("Error accessing %q: %v\n", path, walkErr)
			return walkErr
		}

		for _, skipped := range skip {
			if strings.HasPrefix(path, skipped) {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}
		if info.IsDir() {
			dirName := info.Name()
			if slices.Contains(config.SkippedDirs, dirName) {
				return filepath.SkipDir
			}
		}

		if !info.IsDir() && (filepath.Ext(path) == ".yaml" || filepath.Ext(path) == ".yml" || filepath.Ext(path) == ".json") {
			violations, processErr := processFile(path, requiredTags, caseInsensitive, taggable)
			if processErr != nil {
				log.Printf("Error processing file %s: %v\n", path, processErr)
				return nil // Example: Continue walking
			}
			allViolations = append(allViolations, violations...)
		}
		return nil
	})
	if walkErr != nil {
		log.Printf("Error scanning directory %s: %v\n", dirPath, walkErr)
	}
	return allViolations
}

// processFile parses files and maps the cfn nodes
func processFile(filePath string, requiredTags shared.TagMap, caseInsensitive bool, taggable map[string]bool) ([]shared.Violation, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		log.Printf("Error reading %s: %v\n", filePath, err)
		return nil, fmt.Errorf("reading file %s: %w", filePath, err)
	}
	content := string(data)
	lines := strings.Split(content, "\n")

	skipAll := strings.Contains(content, config.TagNagIgnoreAll)

	root, err := parseYAML(filePath)
	if err != nil {
		return nil, fmt.Errorf("parsing file %s: %w", filePath, err)
	}

	// search root node for resources node
	resourcesMapping := mapNodes(findMapNode(root, "Resources"))
	if resourcesMapping == nil {
		log.Printf("No 'Resources' section found in %s\n", filePath)
		return []shared.Violation{}, nil
	}

	violations := checkResourcesforTags(resourcesMapping, requiredTags, caseInsensitive, lines, skipAll, taggable, filePath)
	return violations, nil
}
