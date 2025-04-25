package cloudformation

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/jakebark/tag-nag/internal/config"
	"github.com/jakebark/tag-nag/internal/shared"
)

// ProcessDirectory walks all cfn files in a directory, then returns violations
func ProcessDirectory(dirPath string, requiredTags map[string][]string, caseInsensitive bool) int {
	var totalViolations int

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Printf("Error accessing %q: %v\n", path, err)
			return err
		}

		if info.IsDir() {
			dirName := info.Name()
			for _, skipped := range config.SkippedDirs {
				if dirName == skipped {
					return filepath.SkipDir
				}
			}
		}

		if !info.IsDir() && (filepath.Ext(path) == ".yaml" || filepath.Ext(path) == ".yml" || filepath.Ext(path) == ".json") {
			violations, processErr := processFile(path, requiredTags, caseInsensitive)
			if processErr != nil {
				log.Printf("Error processing file %s: %v\n", path, processErr)
				return nil // Example: Continue walking
			}
			totalViolations += len(violations)
		}
		return nil
	})
	if err != nil {
		log.Printf("Error scanning directory %s: %v\n", dirPath, err)
	}
	return totalViolations
}

// processFile parses files and maps the cfn nodes
func processFile(filePath string, requiredTags shared.TagMap, caseInsensitive bool) ([]Violation, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Printf("Error reading %s: %v\n", filePath, err)
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
		return []Violation{}, nil
	}

	violations := checkResourcesforTags(resourcesMapping, requiredTags, caseInsensitive, lines, skipAll)

	if len(violations) > 0 {
		fmt.Printf("\nViolation(s) in %s\n", filePath)
		for _, v := range violations {
			if v.skip {
				fmt.Printf("  %d: %s \"%s\" skipped\n", v.line, v.resourceType, v.resourceName)
			} else {
				fmt.Printf("  %d: %s \"%s\" 🏷️  Missing tags: %s\n", v.line, v.resourceType, v.resourceName, strings.Join(v.missingTags, ", "))
			}
		}
	}

	var filteredViolations []Violation
	for _, v := range violations {
		if !v.skip {
			filteredViolations = append(filteredViolations, v)
		}
	}
	return filteredViolations, nil
}
