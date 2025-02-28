package terraform

import (
	"strings"

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
	return normalizeProviderID(providerName, alias, caseInsensitive)
}

func normalizeProviderID(providerName, alias string, caseInsensitive bool) string {
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
			var tags map[string]bool
			if attr, exists := subBlock.Body.Attributes["tags"]; exists {
				var err error
				tags, err = getTagMap(attr, caseInsensitive)
				if err != nil {
					tags = make(map[string]bool)

				}
				// merge resolved tags from locals/vars.
				resolvedTags := resolveDefaultTagReferences(attr, localsAndVars, caseInsensitive)
				tags = mergeTags(tags, resolvedTags)
			} else {
				tags = make(map[string]bool)
			}
			return tags
		}
	}
	return nil
}

func resolveDefaultTagReferences(attr *hclsyntax.Attribute, localsAndVars map[string]map[string]bool, caseInsensitive bool) map[string]bool {
	tagRef := extractTraversalString(attr.Expr, caseInsensitive)
	return localsAndVars[tagRef]
}
