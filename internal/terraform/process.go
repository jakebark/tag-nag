package terraform

import (
	"log"
	"os"
	"path/filepath"
	"slices"
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

	// single directory walk
	tfFiles, err := collectFiles(dirPath, skip)
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
	for _, tf := range tfFiles {
		violations := processFile(tf.path, requiredTags, &defaultTags, tfCtx, caseInsensitive, taggable)
		totalViolations += len(violations)
	}

	return totalViolations
}

// collectFiles identifies all elligible terraform files
func collectFiles(dirPath string, skip []string) ([]tfFile, error) {
	var tfFiles []tfFile

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if skipDirectories(path, info, skip) {
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

// skipDir identifies directories to ignore
func skipDirectories(path string, info os.FileInfo, skip []string) bool {
	// user-defined skip paths
	for _, skipped := range skip {
		if strings.HasPrefix(path, skipped) {
			return true
		}
	}

	// default skipped directories eg .git
	if info.IsDir() {
		dirName := info.Name()
		if slices.Contains(config.SkippedDirs, dirName) {
			return true
		}
	}

	return false
}

// processFile parses files looking for resources
func processFile(filePath string, requiredTags shared.TagMap, defaultTags *DefaultTags, tfCtx *TerraformContext, caseInsensitive bool, taggable map[string]bool) []shared.Violation {
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

	violations := checkResourcesForTags(syntaxBody, requiredTags, defaultTags, tfCtx, caseInsensitive, lines, skipAll, taggable, filePath)

	if len(violations) > 0 {
		log.Printf("\nViolation(s) in %s\n", filePath)
		for _, v := range violations {
			if v.Skip {
				log.Printf("  %d: %s \"%s\" skipped\n", v.Line, v.ResourceType, v.ResourceName)
			} else {
				log.Printf("  %d: %s \"%s\" 🏷️  Missing tags: %s\n", v.Line, v.ResourceType, v.ResourceName, strings.Join(v.MissingTags, ", "))
			}
		}
	}

	var filteredViolations []shared.Violation
	for _, v := range violations {
		if !v.Skip {
			filteredViolations = append(filteredViolations, v)
		}
	}
	return filteredViolations
}
