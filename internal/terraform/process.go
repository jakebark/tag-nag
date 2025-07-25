package terraform

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/jakebark/tag-nag/internal/config"
	"github.com/jakebark/tag-nag/internal/shared"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
)

// ProcessDirectory walks all terraform files in directory
func ProcessDirectory(dirPath string, requiredTags map[string][]string, caseInsensitive bool, skip []string) int {
	hasFiles, err := scan(dirPath)
	if err != nil {
		return 0
	}
	if !hasFiles {
		return 0
	}

	// log.Println("Terraform files found\n")
	var totalViolations int

	taggable := loadTaggableResources("registry.terraform.io/hashicorp/aws")
	if taggable == nil {
		// log.Printf("Warning: Failed to load Terraform AWS Provider\nRun 'terraform init' to fix\n")
		// log.Printf("Continuing with limited features ... \n ")
	}

	tfCtx, err := buildTagContext(dirPath)
	if err != nil {
		tfCtx = &TerraformContext{EvalContext: &hcl.EvalContext{Variables: make(map[string]cty.Value), Functions: make(map[string]function.Function)}}
	}

	defaultTags := DefaultTags{
		LiteralTags: make(map[string]shared.TagMap),
	}

	// first pass, default tags
	err = filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
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
			for _, skippedDir := range config.SkippedDirs {
				if dirName == skippedDir {
					return filepath.SkipDir
				}
			}
		}

		if !info.IsDir() && filepath.Ext(path) == ".tf" {
			parser := hclparse.NewParser()
			file, diags := parser.ParseHCLFile(path)
			if diags.HasErrors() || file == nil {
				log.Printf("Error parsing %s during default tag scan: %v\n", path, diags)
				return nil
			}
			syntaxBody, ok := file.Body.(*hclsyntax.Body)
			if !ok {
				log.Printf("Failed to get syntax body for %s\n", path)
				return nil // Continue walking
			}
			processProviderBlocks(syntaxBody, &defaultTags, tfCtx, caseInsensitive)
		}
		return nil
	})

	//  second pass, evaluate tags on resources
	err = filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
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
			for _, skippedDir := range config.SkippedDirs {
				if dirName == skippedDir {
					return filepath.SkipDir
				}
			}
		}

		if info.IsDir() {
			dirName := info.Name()
			for _, skipped := range config.SkippedDirs {
				if dirName == skipped {
					return filepath.SkipDir
				}
			}
		}
		if !info.IsDir() && filepath.Ext(path) == ".tf" {
			violations := processFile(path, requiredTags, &defaultTags, tfCtx, caseInsensitive, taggable)
			totalViolations += len(violations)
		}
		return nil
	})

	if err != nil {
		log.Printf("Error scanning directory %q: %v\n", dirPath, err)
	}

	return totalViolations
}

// processFile parses files looking for resources
func processFile(filePath string, requiredTags shared.TagMap, defaultTags *DefaultTags, tfCtx *TerraformContext, caseInsensitive bool, taggable map[string]bool) []Violation {
	data, err := os.ReadFile(filePath)
	if err != nil {
		log.Printf("Error reading %s: %v\n", filePath, err)
		return nil
	}
	content := string(data)
	lines := strings.Split(content, "\n")

	skipAll := strings.Contains(content, config.TagNagIgnoreAll)

	parser := hclparse.NewParser()
	file, diagnostics := parser.ParseHCLFile(filePath)

	if diagnostics.HasErrors() {
		log.Printf("Error parsing %s: %v\n", filePath, diagnostics)
		return nil
	}

	syntaxBody, ok := file.Body.(*hclsyntax.Body)
	if !ok {
		log.Printf("Parsing failed for %s\n", filePath)
		return nil
	}

	violations := checkResourcesForTags(syntaxBody, requiredTags, defaultTags, tfCtx, caseInsensitive, lines, skipAll, taggable)

	if len(violations) > 0 {
		log.Printf("\nViolation(s) in %s\n", filePath)
		for _, v := range violations {
			if v.skip {
				log.Printf("  %d: %s \"%s\" skipped\n", v.line, v.resourceType, v.resourceName)
			} else {
				log.Printf("  %d: %s \"%s\" 🏷️  Missing tags: %s\n", v.line, v.resourceType, v.resourceName, strings.Join(v.missingTags, ", "))
			}
		}
	}

	var filteredViolations []Violation
	for _, v := range violations {
		if !v.skip {
			filteredViolations = append(filteredViolations, v)
		}
	}
	return filteredViolations
}

// processProviderBlocks extracts any default_tags from providers
func processProviderBlocks(body *hclsyntax.Body, defaultTags *DefaultTags, tfCtx *TerraformContext, caseInsensitive bool) {
	for _, block := range body.Blocks {
		if block.Type == "provider" && len(block.Labels) > 0 {
			providerID := getProviderID(block, caseInsensitive)
			tags := checkForDefaultTags(block, tfCtx, caseInsensitive)

			if len(tags) > 0 {
				var keys []string
				for key := range tags {
					keys = append(keys, key) // remove bool element of tag map
				}
				sort.Strings(keys)
				fmt.Printf("Found Terraform default tags for provider %s: [%v]\n", providerID, strings.Join(keys, ", "))
				defaultTags.LiteralTags[providerID] = tags

			}
		}
	}
}
