package terraform

import (
	"fmt"
	"strings"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/jakebark/tag-nag/internal/shared"
)

// getProviderID  returns  the provider identifier (aws or alias)
func getProviderID(block *hclsyntax.Block, caseInsensitive bool) string {
	providerName := block.Labels[0]
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

// normalize ProviderID combines the provider name and alias ("aws.west"), aligning with resource provider naming
func normalizeProviderID(providerName, alias string, caseInsensitive bool) string {
	providerID := providerName
	if alias != "" {
		providerID += "." + alias
	}

	if caseInsensitive {
		providerID = strings.ToLower(providerID)
	}

	return providerID
}

// checkForDefaultTags returns the default_tags on a provider block.
// It extracts any literal tags (with extractTags) and merges them with any referenced tags
func checkForDefaultTags(block *hclsyntax.Block, referencedTags TagReferences, caseInsensitive bool) shared.TagMap {
	for _, subBlock := range block.Body.Blocks {
		if subBlock.Type == "default_tags" {
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

// resolveDefaultTagReferences looks up referencedTags (locals/vars)
func resolveDefaultTagReferences(attr *hclsyntax.Attribute, referencedTags TagReferences, caseInsensitive bool) shared.TagMap {
	tagRef := traversalToString(attr.Expr, caseInsensitive)
	if tagRef == "" {
		// if the expr isn’t a ScopeTraversalExpr, try formatting it as a string.
		tagRef = strings.TrimSpace(fmt.Sprintf("%v", attr.Expr))
	}
	return referencedTags[tagRef]
}
