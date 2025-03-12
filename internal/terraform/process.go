package terraform

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

// ProcessDirectory walks all terraform files in directory
func ProcessDirectory(dirPath string, requiredTags map[string]string, caseInsensitive bool) int {
	var totalViolations int
	defaultTags := DefaultTags{
		LiteralTags:    make(TagReferences),
		ReferencedTags: checkReferencedTags(dirPath),
	}

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".tf") {
			processProvider(path, &defaultTags, caseInsensitive)
		}
		return nil
	})
	if err != nil {
		fmt.Println("Error finding provider:", err)
	}

	err = filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".tf") {
			violations := processFile(path, requiredTags, &defaultTags, caseInsensitive)
			totalViolations += len(violations)
		}
		return nil
	})
	if err != nil {
		fmt.Println("Error scanning directory:", err)
	}

	return totalViolations
}

// processProvider parses files looking for providers
func processProvider(filePath string, defaultTags *DefaultTags, caseInsensitive bool) {
	parser := hclparse.NewParser()
	file, diagnostics := parser.ParseHCLFile(filePath)

	if diagnostics.HasErrors() {
		fmt.Printf("Error parsing %s: %v\n", filePath, diagnostics)
		return
	}
	syntaxBody, ok := file.Body.(*hclsyntax.Body)
	if !ok {
		return
	}
	processProviderBlocks(syntaxBody, defaultTags, caseInsensitive)
}

// processFile parses files looking for resources
func processFile(filePath string, requiredTags TagMap, defaultTags *DefaultTags, caseInsensitive bool) []Violation {
	parser := hclparse.NewParser()
	file, diagnostics := parser.ParseHCLFile(filePath)

	if diagnostics.HasErrors() {
		fmt.Printf("Error parsing %s: %v\n", filePath, diagnostics)
		return nil
	}

	syntaxBody, ok := file.Body.(*hclsyntax.Body)
	if !ok {
		fmt.Printf("Parsing failed for %s\n", filePath)
		return nil
	}

	// processProviderBlocks(syntaxBody, defaultTags, caseInsensitive)
	violations := processResourceBlocks(syntaxBody, requiredTags, defaultTags, caseInsensitive)

	if len(violations) > 0 {
		fmt.Printf("\nViolation(s) in %s\n", filePath)
		for _, v := range violations {
			fmt.Printf("  %d: %s \"%s\" 🏷️  Missing tags: %s\n", v.line, v.resourceType, v.resourceName, strings.Join(v.missingTags, ", "))
		}
	}
	return violations
}

// processProviderBlocks extracts any default_tags from providers
func processProviderBlocks(body *hclsyntax.Body, defaultTags *DefaultTags, caseInsensitive bool) {
	for _, block := range body.Blocks {
		if block.Type == "provider" && len(block.Labels) > 0 {
			providerID := getProviderID(block, caseInsensitive)
			tags := checkForDefaultTags(block, defaultTags.ReferencedTags, caseInsensitive)

			if len(tags) > 0 {
				var keys []string
				for key := range tags {
					keys = append(keys, key) // remove bool element of tag map
				}
				fmt.Printf("Found Terraform default tags for provider %s: [%v]\n", providerID, strings.Join(keys, ", "))

			}
			if tags != nil {
				defaultTags.LiteralTags[providerID] = tags
			}
		}
	}
}

// processResourceBlocks initiates checking a resource for tags
func processResourceBlocks(body *hclsyntax.Body, requiredTags TagMap, defaultTags *DefaultTags, caseInsensitive bool) []Violation {
	return checkResourcesForTags(body, requiredTags, defaultTags, caseInsensitive)
}
