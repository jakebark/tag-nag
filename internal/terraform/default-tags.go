package terraform

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

func getProviderID(block *hclsyntax.Block, caseInsensitive bool) string {
	providerName := block.Labels[0] // providers have one name: provider "aws"
	var alias string

	// check for alias presence
	if attr, ok := block.Body.Attributes["alias"]; ok {
		val, diags := attr.Expr.Value(nil)
		if !diags.HasErrors() {
			alias = val.AsString()
		}
	}

	// providerID combines provider name and alias ("aws.west") to align with resource provider arg
	providerID := providerName
	if alias != "" {
		providerID += "." + alias
	}

	if caseInsensitive {
		providerID = strings.ToLower(providerID)
	}

	return providerID
}

func checkDefaultTags(block *hclsyntax.Block, localsAndVars map[string]map[string]bool, caseInsensitive bool) map[string]bool {
	for _, subBlock := range block.Body.Blocks {
		if subBlock.Type == "default_tags" {
			tags, _ := findTags(subBlock, caseInsensitive) // call findTags to extract default_tags

			// if findTags fails, initialize as an empty map to prevent error
			if tags == nil {
				tags = make(map[string]bool)
			}

			// if tags are a var/local
			if attr, exists := subBlock.Body.Attributes["tags"]; exists {
				resolvedTags := resolveDefaultTagReferences(attr, localsAndVars)
				for k, v := range resolvedTags {
					tags[k] = v
				}
			}
			return tags
		}
	}
	return nil
}

func resolveDefaultTagReferences(attr *hclsyntax.Attribute, localsAndVars map[string]map[string]bool) map[string]bool {
	traversalExpr, ok := attr.Expr.(*hclsyntax.ScopeTraversalExpr) // ScopeTraversalExpr == value is a reference, not a direct map
	if !ok {
		return nil
	}

	// converts tags = local.tags to ... local.tags
	// sounds dumb, but hcl sees it as hierachical components, eg:
	// TraversalExpr{
	// Traversal: [
	//    TraverseRoot{Name: "local"},
	//    TraverseAttr{Name: "tags"}
	// ]}

	refParts := make([]string, len(traversalExpr.Traversal))
	for i, step := range traversalExpr.Traversal {
		switch t := step.(type) {
		case hcl.TraverseRoot:
			refParts[i] = t.Name // local
		case hcl.TraverseAttr:
			refParts[i] = t.Name // tags
		}
	} // ["local", "tags"]
	ref := strings.Join(refParts, ".") // local.tags

	return localsAndVars[ref]
}

func checkVariables(dirPath string) map[string]map[string]bool {
	localsAndVars := make(map[string]map[string]bool)

	_ = filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".tf") {
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
					for name, attr := range block.Body.Attributes {
						val, diags := attr.Expr.Value(nil)
						if !diags.HasErrors() && val.Type().IsObjectType() {
							localTags := make(map[string]bool)
							for key := range val.AsValueMap() {
								localTags[key] = true
							}
							localsAndVars["local."+name] = localTags
						}
					}
				}

				if block.Type == "variable" && len(block.Labels) > 0 {
					varName := block.Labels[0]
					if attr, ok := block.Body.Attributes["default"]; ok {
						val, diags := attr.Expr.Value(nil)
						if !diags.HasErrors() && val.Type().IsObjectType() {
							varTags := make(map[string]bool)
							for key := range val.AsValueMap() {
								varTags[key] = true
							}
							localsAndVars["var."+varName] = varTags
						}
					}
				}
			}
		}
		return nil
	})

	return localsAndVars
}
