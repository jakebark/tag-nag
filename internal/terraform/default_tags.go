package terraform

import (
	"strings"

	"github.com/hashicorp/hcl/v2"
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
			refParts[i] = t.Name
		case hcl.TraverseAttr:
			refParts[i] = t.Name
		}
	}
	tagRef := strings.Join(refParts, ".")

	return localsAndVars[tagRef]
}
