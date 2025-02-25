package terraform

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

func ScanDirectory(dirPath string, requiredTags []string, caseInsensitive bool) {
	defaultTagsByProvider := findDefaultTagsByProvider(dirPath, caseInsensitive)

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".tf") {
			checkTerraformFile(path, requiredTags, defaultTagsByProvider, caseInsensitive)
		}
		return nil
	})
	if err != nil {
		fmt.Println("Error scanning directory:", err)
	}
}

func checkTerraformFile(filePath string, requiredTags []string, defaultTagsByProvider map[string]map[string]bool, caseInsensitive bool) {
	parser := hclparse.NewParser()
	file, diags := parser.ParseHCLFile(filePath)
	if diags.HasErrors() {
		fmt.Printf("Error parsing %s: %v\n", filePath, diags)
		return
	}

	syntaxBody, ok := file.Body.(*hclsyntax.Body)
	if !ok {
		fmt.Printf("Parsing failed for %s\n", filePath)
		return
	}

	violations := checkResources(syntaxBody, requiredTags, defaultTagsByProvider, caseInsensitive)
	if len(violations) > 0 {
		// Find the earliest violation line to use in the header.
		headerLine := violations[0].line
		for _, v := range violations {
			if v.line < headerLine {
				headerLine = v.line
			}
		}
		fmt.Printf("\nNon-compliant resources in %s (line %d)\n", filePath, headerLine)
		for _, v := range violations {
			fmt.Printf("  %s \"%s\" (line %d)\n", v.resourceType, v.resourceName, v.line)
			fmt.Printf("  ❌ Missing tags: %s\n", strings.Join(v.missingTags, ", "))
		}
	}
}
