package terraform

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

// checkReferencedTags looks for locals and vars in the directory, then returns a map of them
func checkReferencedTags(dirPath string) TagReferences {
	referencedTags := make(TagReferences)

	_ = filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".tf") {
			return nil
		}

		parser := hclparse.NewParser()
		file, diags := parser.ParseHCLFile(path)
		if diags.HasErrors() {
			return nil
		}

		syntaxBody, ok := file.Body.(*hclsyntax.Body)
		if !ok {
			return nil
		}

		for _, block := range syntaxBody.Blocks {
			if block.Type == "locals" {
				extractLocals(block, referencedTags)
			} else if block.Type == "variable" {
				extractVariables(block, referencedTags)
			}
		}
		return nil
	})

	return referencedTags
}

// extractLocals extracts hcl tag maps from locals (using extractTags) and appends them to the defaultTags struct (defaultTags.referencedTags)
func extractLocals(block *hclsyntax.Block, referencedTags TagReferences) {
	for name, attr := range block.Body.Attributes {
		tags := extractTags(attr, false)
		referencedTags["local."+name] = tags
	}
}

// extractVariables extracts hcl tag maps from vars (using extractTags) and appends them to the defaultTags struct (defaultTags.referencedTags)
func extractVariables(block *hclsyntax.Block, referencedTags TagReferences) {
	if len(block.Labels) > 0 {
		if attr, ok := block.Body.Attributes["default"]; ok {
			tags := extractTags(attr, false)
			referencedTags["var."+block.Labels[0]] = tags
		}
	}
}
