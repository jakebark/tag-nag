package terraform

import (
	"os"
	"path/filepath"

	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/jakebark/tag-nag/internal/shared"
	"github.com/zclconf/go-cty/cty"
)

// checkReferencedTags looks for locals and vars in the directory, then returns a map of them
func checkReferencedTags(dirPath string) TagReferences {
	referencedTags := make(TagReferences)

	_ = filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || filepath.Ext(path) != ".tf" {
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
		if v, diags := attr.Expr.Value(nil); !diags.HasErrors() && v.Type().Equals(cty.String) {
			referencedTags["local."+name] = shared.TagMap{"_": {v.AsString()}}
		} else {
			tags := extractTags(attr, false)
			referencedTags["local."+name] = tags
		}
	}
}

// extractVariables extracts hcl tag maps from vars (using extractTags) and appends them to the defaultTags struct (defaultTags.referencedTags)
func extractVariables(block *hclsyntax.Block, referencedTags TagReferences) {
	if len(block.Labels) > 0 {
		if attr, ok := block.Body.Attributes["default"]; ok {
			if v, diags := attr.Expr.Value(nil); !diags.HasErrors() && v.Type().Equals(cty.String) {
				referencedTags["var."+block.Labels[0]] = shared.TagMap{"_": {v.AsString()}}
			} else {
				tags := extractTags(attr, false)
				referencedTags["var."+block.Labels[0]] = tags
			}
		}
	}
}
