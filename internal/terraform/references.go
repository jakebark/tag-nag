package terraform

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

func checkReferences(dirPath string) map[string]map[string]bool {
	refs := make(map[string]map[string]bool)

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
				extractLocals(block, refs)
			} else if block.Type == "variable" {
				extractVariables(block, refs)
			}
		}
		return nil
	})

	return refs
}

func extractLocals(block *hclsyntax.Block, refs map[string]map[string]bool) {
	for name, attr := range block.Body.Attributes {
		if tags, err := getTagMap(attr, false); err == nil {
			refs["local."+name] = tags
		}
	}
}

func extractVariables(block *hclsyntax.Block, refs map[string]map[string]bool) {
	if len(block.Labels) > 0 {
		if attr, ok := block.Body.Attributes["default"]; ok {
			if tags, err := getTagMap(attr, false); err == nil {
				refs["var."+block.Labels[0]] = tags
			}
		}
	}
}
