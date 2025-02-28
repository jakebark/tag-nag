package terraform

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

type DefaultTags struct {
	ProviderTags map[string]map[string]bool
	References   map[string]map[string]bool
}

// check file(s) and presence of default tags
func ProcessDirectory(dirPath string, requiredTags []string, caseInsensitive bool) {
	defaultTags := DefaultTags{
		ProviderTags: make(map[string]map[string]bool),
		References:   checkReferences(dirPath),
	}

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".tf") { // if .tf file, assess with checkFile
			processFile(path, requiredTags, &defaultTags, caseInsensitive)
		}
		return nil
	})
	if err != nil {
		fmt.Println("Error scanning directory:", err)
	}
}

func processFile(filePath string, requiredTags []string, defaultTags *DefaultTags, caseInsensitive bool) {
	parser := hclparse.NewParser()
	file, diagnostics := parser.ParseHCLFile(filePath)

	//we dont  parse invalid tf
	if diagnostics.HasErrors() {
		fmt.Printf("Error parsing %s: %v\n", filePath, diagnostics)
		return
	}

	// convert file into  stuctured hclsyntax.body
	syntaxBody, ok := file.Body.(*hclsyntax.Body)
	if !ok {
		fmt.Printf("Parsing failed for %s\n", filePath)
		return
	}

	processProviderBlocks(syntaxBody, defaultTags, caseInsensitive) // kick off func to update defaultTags struct (defaultTags.ProviderTags)
	violations := processResourceBlocks(syntaxBody, requiredTags, defaultTags, caseInsensitive)

	if len(violations) > 0 {
		fmt.Printf("\nNon-compliant resources in %s\n", filePath)
		for _, v := range violations {
			fmt.Printf("  %s \"%s\" (line %d)\n", v.resourceType, v.resourceName, v.line)
			fmt.Printf("  🏷️ Missing tags: %s\n", strings.Join(v.missingTags, ", "))
		}
	}
}

func processProviderBlocks(body *hclsyntax.Body, defaultTags *DefaultTags, caseInsensitive bool) {
	for _, block := range body.Blocks {
		if block.Type == "provider" && len(block.Labels) > 0 {
			providerID := getProviderID(block, caseInsensitive)                         // providerID combines provider name and alias ("aws.west") to align with resource provider arg
			tags := checkforDefaultTags(block, defaultTags.References, caseInsensitive) // check provider for default_tags, return map of tags

			if len(tags) > 0 {
				fmt.Printf("🔍 Found default tags for provider %s: %v\n", providerID, tags)
			}
			if tags != nil {
				defaultTags.ProviderTags[providerID] = tags
			}
		}
	}
}

func processResourceBlocks(body *hclsyntax.Body, requiredTags []string, defaultTags *DefaultTags, caseInsensitive bool) []Violation {
	return checkResourcesForTags(body, requiredTags, defaultTags, caseInsensitive)
}
