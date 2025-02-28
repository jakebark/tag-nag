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
	providerID := providerName
	if alias != "" {
		providerID += "." + alias // combines provider name and alias ("aws.west") to align with resource provider arg
	}

	if caseInsensitive {
		providerID = strings.ToLower(providerID)
	}

	return providerID
}

func checkForDefaultTags(block *hclsyntax.Block, referencedTags map[string]map[string]bool, caseInsensitive bool) map[string]bool {
	for _, subBlock := range block.Body.Blocks {
		if subBlock.Type == "default_tags" {
			// tags sub-block always exists, dont need to "if exists"
			attr := subBlock.Body.Attributes["tags"]

			tags := extractTags(attr, caseInsensitive)

			// merge literal tags (above) with referenced tags (locals, vars)
			resolved := resolveDefaultTagReferences(attr, referencedTags, caseInsensitive)
			tags = mergeTags(tags, resolved)

			return tags
		}
	}
	return nil
}
func resolveDefaultTagReferences(attr *hclsyntax.Attribute, referencedTags map[string]map[string]bool, caseInsensitive bool) map[string]bool {
	tagRef := extractTraversalString(attr.Expr, caseInsensitive)
	return referencedTags[tagRef]
}
