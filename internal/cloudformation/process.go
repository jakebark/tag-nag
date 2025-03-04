package cloudformation

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// ProcessDirectory parses cfn files and returns the total amount of violations found
func ProcessDirectory(dirPath string, requiredTags []string, caseInsensitive bool) int {
	var totalViolations int

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && (strings.HasSuffix(path, ".yaml") || strings.HasSuffix(path, ".yml") || strings.HasSuffix(path, ".json")) {
			violations, err := processFile(path, requiredTags, caseInsensitive)
			if err != nil {
				fmt.Printf("Error processing file %s: %v\n", path, err)
			}
			totalViolations += len(violations)
		}
		return nil
	})
	if err != nil {
		fmt.Println("Error scanning directory:", err)
	}
	return totalViolations
}

// processFile parses cloudformation files and returns any violations
func processFile(filePath string, requiredTags []string, caseInsensitive bool) ([]Violation, error) {
	root, err := parseYAML(filePath)
	if err != nil {
		return nil, err
	}

	// search root node for resources node
	resourcesMapping := mapNodes(findMapNode(root, "Resources"))
	if resourcesMapping == nil {
		return []Violation{}, nil
	}

	violations := processResourceBlocks(resourcesMapping, requiredTags, caseInsensitive)

	if len(violations) > 0 {
		fmt.Printf("\nViolation(s) in %s\n", filePath)
		for _, v := range violations {
			fmt.Printf("  %d: %s \"%s\" 🏷️  Missing tags: %v\n", v.Line, v.ResourceType, v.ResourceName, v.MissingTags)
		}
	}

	return violations, nil
}

// processResourceBlocks initiates checking a resource for tags
func processResourceBlocks(resourcesMapping map[string]*yaml.Node, requiredTags []string, caseInsensitive bool) []Violation {
	return getResourceViolations(resourcesMapping, requiredTags, caseInsensitive)
}
