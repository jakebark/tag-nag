package terraform

import (
	"log"
	"strings"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/jakebark/tag-nag/internal/shared"
	"github.com/zclconf/go-cty/cty"
)

// checkResourcesForTags inspects resource blocks and returns violations
func checkResourcesForTags(body *hclsyntax.Body, requiredTags shared.TagMap, defaultTags *DefaultTags, tfCtx *TerraformContext, caseInsensitive bool, fileLines []string, skipAll bool, taggable map[string]bool, filePath string) []shared.Violation {
	var violations []shared.Violation

	for _, block := range body.Blocks {
		if block.Type != "resource" || len(block.Labels) < 2 { // skip anything without 2 labels eg "aws_s3_bucket" and "this"
			continue
		}

		resourceType := block.Labels[0] // aws_s3_bucket
		resourceName := block.Labels[1] // this

		if !strings.HasPrefix(resourceType, "aws_") {
			continue
		}

		isTaggable := true // assume resource is taggable, by default
		if taggable != nil {
			var found bool
			isTaggable, found = taggable[resourceType]
			if !found {
				isTaggable = true // if not found, assume resource is taggable
				// isTaggable = false
				// log.Printf("Warning: Resource type %s not found in provider schema. Assuming not taggable.", resourceType) //todo
			}
		} else {
		}

		if !isTaggable {
			// log.Printf("Skipping non-taggable resource type: %s", resourceType)
			continue
		}

		providerID := getResourceProvider(block, caseInsensitive)
		providerEvalTags := defaultTags.LiteralTags[providerID]
		if providerEvalTags == nil {
			providerEvalTags = make(shared.TagMap)
		}

		resourceEvalTags := findTags(block, tfCtx, caseInsensitive)
		effectiveTags := mergeTags(providerEvalTags, resourceEvalTags)

		missingTags := shared.FilterMissingTags(requiredTags, effectiveTags, caseInsensitive)
		if len(missingTags) > 0 {
			violation := shared.Violation{
				ResourceType: resourceType,
				ResourceName: resourceName,
				Line:         block.DefRange().Start.Line,
				MissingTags:  missingTags,
				FilePath:     filePath,
			}
			if skipAll || SkipResource(block, fileLines) {
				violation.Skip = true
			}
			violations = append(violations, violation)
		}
	}
	return violations
}

// getResourceProvider determines the provider for a resource block
func getResourceProvider(block *hclsyntax.Block, caseInsensitive bool) string {
	if attr, ok := block.Body.Attributes["provider"]; ok {

		// provider is a literal string ("aws")
		val, diags := attr.Expr.Value(nil)
		if !diags.HasErrors() {
			s := val.AsString()
			if caseInsensitive {
				s = strings.ToLower(s)
			}
			return s
		}
		// provider is not a literal string ("aws.west")
		s := traversalToString(attr.Expr, caseInsensitive)
		if s != "" {
			return s
		}
	}

	// no explicit provider, return default provider
	defaultProvider := "aws"
	if caseInsensitive {
		defaultProvider = strings.ToLower(defaultProvider)
	}
	return defaultProvider
}

// findTags returns tag map from a resource block (with extractTags), if it has tags
func findTags(block *hclsyntax.Block, tfCtx *TerraformContext, caseInsensitive bool) shared.TagMap {
	evalTags := make(shared.TagMap)
	if attr, exists := block.Body.Attributes["tags"]; exists {

		childCtx := tfCtx.EvalContext.NewChild()
		if childCtx.Variables == nil {
			childCtx.Variables = make(map[string]cty.Value)
		}
		childCtx.Variables["each"] = cty.ObjectVal(map[string]cty.Value{
			"key":   cty.StringVal(""),
			"value": cty.StringVal(""),
		})
		tagsVal, diags := attr.Expr.Value(childCtx)

		if diags.HasErrors() {
			log.Printf("Error evaluating tags for resource %s.%s: %v", block.Labels[0], block.Labels[1], diags)
			return evalTags
		}

		if !tagsVal.Type().IsObjectType() && !tagsVal.Type().IsMapType() {
			return evalTags
		}
		if tagsVal.IsNull() {
			return evalTags
		}

		for key, val := range tagsVal.AsValueMap() {
			var valStr string
			if val.IsNull() {
				valStr = ""
			} else if val.Type() == cty.String {
				valStr = val.AsString()
			} else {
				strResult, err := convertCtyValueToString(val)
				if err != nil {
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
	}
	return evalTags
}
