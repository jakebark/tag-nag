package terraform

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

// DefaultTags hold the default_tag values
type DefaultTags struct {
	LiteralTags    map[string]map[string]bool
	ReferencedTags map[string]map[string]bool
}

// ProcessDirectory identifies all terraform files in directory
func ProcessDirectory(dirPath string, requiredTags []string, caseInsensitive bool) {
	defaultTags := DefaultTags{
		LiteralTags:    make(map[string]map[string]bool),
		ReferencedTags: checkReferencedTags(dirPath),
	}

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".tf") {
			processFile(path, requiredTags, &defaultTags, caseInsensitive)
		}
		return nil
	})
	if err != nil {
		fmt.Println("Error scanning directory:", err)
	}
}

// processFile parses terraform files.
// It checks all providers and updates the defaultTags struct (processProviderBlocks).
// Then it returns a list of violations with (processResourceBlocks)
func processFile(filePath string, requiredTags []string, defaultTags *DefaultTags, caseInsensitive bool) {
	parser := hclparse.NewParser()
	file, diagnostics := parser.ParseHCLFile(filePath)

	if diagnostics.HasErrors() {
		fmt.Printf("Error parsing %s: %v\n", filePath, diagnostics)
		return
	}

	syntaxBody, ok := file.Body.(*hclsyntax.Body)
	if !ok {
		fmt.Printf("Parsing failed for %s\n", filePath)
		return
	}

	processProviderBlocks(syntaxBody, defaultTags, caseInsensitive)
	violations := processResourceBlocks(syntaxBody, requiredTags, defaultTags, caseInsensitive)

	if len(violations) > 0 {
		fmt.Printf("\nNon-compliant resources in %s\n", filePath)
		for _, v := range violations {
			fmt.Printf("  %s \"%s\" (line %d)\n", v.resourceType, v.resourceName, v.line)
			fmt.Printf("  🏷️ Missing tags: %s\n", strings.Join(v.missingTags, ", "))
		}
	}
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
				fmt.Printf("🔍 Found default tags for provider %s: %v\n", providerID, keys)

			}
			if tags != nil {
				defaultTags.LiteralTags[providerID] = tags
			}
		}
	}
}

// processResourceBlocks initiates chekcing a resource for tags
func processResourceBlocks(body *hclsyntax.Body, requiredTags []string, defaultTags *DefaultTags, caseInsensitive bool) []Violation {
	return checkResourcesForTags(body, requiredTags, defaultTags, caseInsensitive)
}
