package terraform

import (
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/jakebark/tag-nag/internal/shared"
	"github.com/zclconf/go-cty/cty"
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
func checkForDefaultTags(block *hclsyntax.Block, tfCtx *TerraformContext, caseInsensitive bool) shared.TagMap { // Add tfCtx param
	for _, subBlock := range block.Body.Blocks {
		if subBlock.Type == "default_tags" {
			if tagsAttr, exists := subBlock.Body.Attributes["tags"]; exists {
				tagsVal, diags := tagsAttr.Expr.Value(tfCtx.EvalContext)
				if diags.HasErrors() {
					log.Printf("Error evaluating default_tags expression for provider %v: %v", block.Labels, diags)
					return nil
				}

				if !tagsVal.Type().IsObjectType() && !tagsVal.Type().IsMapType() {
					log.Printf("Warning: Evaluated default_tags for provider %v is not an object/map type, but %s. Skipping.", block.Labels, tagsVal.Type().FriendlyName())
					return nil
				}
				if tagsVal.IsNull() {
					log.Printf("Warning: Evaluated default_tags for provider %v is null. Skipping.", block.Labels)
					return nil
				}

				evalTags := make(shared.TagMap)
				for key, val := range tagsVal.AsValueMap() {
					var valStr string
					if val.IsNull() {
						valStr = ""
					} else if val.Type() == cty.String {
						valStr = val.AsString()
					} else {
						strResult, err := convertCtyValueToString(val)
						if err != nil {
							log.Printf("Warning: Could not convert default tag value for key %q to string: %v. Using empty string.", key, err)
							valStr = ""
						} else {
							valStr = strResult
						}
					}

					effectiveKey := key
					if caseInsensitive {
						effectiveKey = strings.ToLower(key)
					}
					evalTags[effectiveKey] = []string{valStr}
				}
				return evalTags
			}
		}
	}
	return nil // No default_tags block found
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
