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

type tfFile struct {
	path string
	info os.FileInfo
}

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

	// testing single directory walk
	tfFiles, err := collectTerraformFiles(dirPath, skip)
	if err != nil {
		log.Printf("Error scanning directory %q: %v\n", dirPath, err)
		return 0
	}

	if len(tfFiles) == 0 {
		return 0
	}

	// extract default tags from all files
	defaultTags := processDefaultTags(tfFiles, tfCtx, caseInsensitive)

	// process resources for tag violations
	totalViolations = processResourceViolations(tfFiles, requiredTags, defaultTags, tfCtx, caseInsensitive, taggable)

	return totalViolations
}

func collectTerraformFiles(dirPath string, skip []string) ([]tfFile, error) {
	var tfFiles []tfFile

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if shouldSkipPath(path, info, skip) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if !info.IsDir() && filepath.Ext(path) == ".tf" {
			tfFiles = append(tfFiles, tfFile{path: path, info: info})
		}
		return nil
	})

	return tfFiles, err
}

func shouldSkipPath(path string, info os.FileInfo, skip []string) bool {
	// user-defined skip paths
	for _, skipped := range skip {
		if strings.HasPrefix(path, skipped) {
			return true
		}
	}

	// default skipped directories eg .git
	if info.IsDir() {
		dirName := info.Name()
		for _, skippedDir := range config.SkippedDirs {
			if dirName == skippedDir {
				return true
			}
		}
	}

	return false
}

func processDefaultTags(tfFiles []tfFile, tfCtx *TerraformContext, caseInsensitive bool) DefaultTags {
	defaultTags := DefaultTags{
		LiteralTags: make(map[string]shared.TagMap),
	}

	for _, tf := range tfFiles {
		parser := hclparse.NewParser()
		file, diags := parser.ParseHCLFile(tf.path)
		if diags.HasErrors() || file == nil {
			log.Printf("Error parsing %s during default tag scan: %v\n", tf.path, diags)
			continue
		}

		syntaxBody, ok := file.Body.(*hclsyntax.Body)
		if !ok {
			log.Printf("Failed to get syntax body for %s\n", tf.path)
			continue
		}

		processProviderBlocks(syntaxBody, &defaultTags, tfCtx, caseInsensitive)
	}

	return defaultTags
}

func processResourceViolations(tfFiles []tfFile, requiredTags shared.TagMap, defaultTags DefaultTags, tfCtx *TerraformContext, caseInsensitive bool, taggable map[string]bool) int {
	var totalViolations int

	for _, tf := range tfFiles {
		violations := processFile(tf.path, requiredTags, &defaultTags, tfCtx, caseInsensitive, taggable)
		totalViolations += len(violations)
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
