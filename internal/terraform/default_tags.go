package terraform

import (
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/jakebark/tag-nag/internal/shared"
	"github.com/zclconf/go-cty/cty"
)

// processDefaultTags identifies the default tags
func processDefaultTags(tfFiles []tfFile, tfCtx *TerraformContext, caseInsensitive bool) DefaultTags {
	defaultTags := DefaultTags{
		LiteralTags: make(map[string]shared.TagMap),
	}

	parser := hclparse.NewParser()

	for _, tf := range tfFiles {
		file, diags := parser.ParseHCLFile(tf.path)
		if diags.HasErrors() || file == nil {
			log.Printf("Error parsing %s during default tag scan: %v\n", tf.path, diags)
			continue
		}

		syntaxBody, ok := file.Body.(*hclsyntax.Body)
		if !ok {
			log.Printf("Failed to get syntax body for %s\n", tf.path)
			continue
		}

		processProviders(syntaxBody, &defaultTags, tfCtx, caseInsensitive)
	}

	return defaultTags
}

// processProviders extracts any default_tags from providers
func processProviders(body *hclsyntax.Body, defaultTags *DefaultTags, tfCtx *TerraformContext, caseInsensitive bool) {
	for _, block := range body.Blocks {
		if block.Type == "provider" && len(block.Labels) > 0 {
			providerID := getProviderID(block, caseInsensitive)   // handle ID
			tags := getDefaultTags(block, tfCtx, caseInsensitive) // handle tags

			if len(tags) > 0 {
				var keys []string
				for key := range tags {
					keys = append(keys, key) // remove bool element of tag map
				}
				sort.Strings(keys)
				fmt.Printf("Found Terraform default tags for provider %s: [%v]\n", providerID, strings.Join(keys, ", "))
				defaultTags.LiteralTags[providerID] = tags

			}
		}
	}
}

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

// getDefaultTags returns the default_tags on a provider block.
func getDefaultTags(block *hclsyntax.Block, tfCtx *TerraformContext, caseInsensitive bool) shared.TagMap { // Add tfCtx param
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
