package cloudformation

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

var (
	ignoreAllRegex = regexp.MustCompile(`(?m)^(#|//)tag:nag ignore-all\b`)
	ignoreRegex    = regexp.MustCompile(`(?m)^(#|//)tag:nag ignore\b`)
)

// ProcessDirectory walks all cfn files in a directory, then returns violations
func ProcessDirectory(dirPath string, requiredTags map[string]string, caseInsensitive bool) int {
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

// processFile parses files and maps the cfn nodes
func processFile(filePath string, requiredTags map[string]string, caseInsensitive bool) ([]Violation, error) {

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	fileText := string(data)

	// Check for file-level ignore-all using the regex.
	if ignoreAllRegex.MatchString(fileText) {
		fmt.Printf("Skipping %s\n", filePath)
		return nil, nil
	}

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
	var skipped []Violation
	violations, skipped = skipResources(violations, fileText)

	if len(violations) > 0 {
		fmt.Printf("\nViolation(s) in %s\n", filePath)
		for _, v := range violations {
			fmt.Printf("  %d: %s \"%s\" 🏷️  Missing tags: %v\n", v.Line, v.ResourceType, v.ResourceName, strings.Join(v.MissingTags, ", "))
		}
	}

	if len(skipped) > 0 {
		fmt.Printf("\nResource skip(s) in %s:\n", filePath)
		for _, v := range skipped {
			fmt.Printf("  %d: %s \"%s\"\n", v.Line, v.ResourceType, v.ResourceName)
		}
	}

	return violations, nil
}

// processResourceBlocks initiates checking a resource for tags
func processResourceBlocks(resourcesMapping map[string]*yaml.Node, requiredTags TagMap, caseInsensitive bool) []Violation {
	return checkResourcesforTags(resourcesMapping, requiredTags, caseInsensitive)
}
