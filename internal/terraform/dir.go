package terraform

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

// check file(s) and presence of default tags
func ScanDirectory(dirPath string, requiredTags []string, caseInsensitive bool) {
	defaultTags := checkForDefaultTags(dirPath, caseInsensitive)

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".tf") { // if .tf file, assess with checkTerraformFile
			checkFile(path, requiredTags, defaultTags, caseInsensitive)
		}
		return nil
	})
	if err != nil {
		fmt.Println("Error scanning directory:", err)
	}
}

func checkFile(filePath string, requiredTags []string, defaultTagsByProvider map[string]map[string]bool, caseInsensitive bool) {
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

	// feed all into checkResources to check invidivual resources
	// violations = a resource missing requiredTags
	violations := checkResources(syntaxBody, requiredTags, defaultTagsByProvider, caseInsensitive)
	if len(violations) > 0 {
		fmt.Printf("\nNon-compliant resources in %s\n", filePath)
		for _, v := range violations {
			fmt.Printf("  %s \"%s\" (line %d)\n", v.resourceType, v.resourceName, v.line)
			fmt.Printf("  ❌ Missing tags: %s\n", strings.Join(v.missingTags, ", "))
		}
	}
}
