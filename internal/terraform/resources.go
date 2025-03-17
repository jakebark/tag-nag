package terraform

import (
	"fmt"
	"strings"

	"github.com/hashicorp/hcl/v2/hclsyntax"
)

// checkResourcesForTags inspects resource blocks and returns violations
func checkResourcesForTags(body *hclsyntax.Body, requiredTags TagMap, defaultTags *DefaultTags, caseInsensitive bool, fileLines []string, skipAll bool) []Violation {
	var violations []Violation

	for _, block := range body.Blocks {
		if block.Type != "resource" || len(block.Labels) < 2 { // skip anything without 2 labels eg "aws_s3_bucket" and "this"
			continue
		}

		resourceType := block.Labels[0] // aws_s3_bucket
		resourceName := block.Labels[1] // this

		if !strings.HasPrefix(resourceType, "aws_") {
			continue
		}

		providerID := getResourceProvider(block, caseInsensitive)
		providerLiteralTags := defaultTags.LiteralTags[providerID]
		if providerLiteralTags == nil {
			providerLiteralTags = make(TagMap)
		}

		resourceTags := findTags(block, defaultTags.ReferencedTags, caseInsensitive)
		effectiveTags := mergeTags(providerLiteralTags, resourceTags)

		missingTags := filterMissingTags(requiredTags, effectiveTags, caseInsensitive)
		if len(missingTags) > 0 {
			violation := Violation{
				resourceType: resourceType,
				resourceName: resourceName,
				line:         block.DefRange().Start.Line,
				missingTags:  missingTags,
			}
			// Set skip flag if file-level or resource-level ignore is found.
			if skipAll || skipResource(block, fileLines) {
				violation.skip = true
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
func findTags(block *hclsyntax.Block, referencedTags TagReferences, caseInsensitive bool) TagMap {
	if attr, ok := block.Body.Attributes["tags"]; ok {
		// literal tags
		tags := extractTags(attr, caseInsensitive)
		if len(tags) > 0 {
			return tags
		}

		// referenced tags
		refKey := traversalToString(attr.Expr, caseInsensitive)
		if refKey != "" {
			if resolved, ok := referencedTags[refKey]; ok {
				return resolved
			}
		}
	}
	return make(TagMap)
}

func filterMissingTags(requiredTags TagMap, effectiveTags TagMap, caseInsensitive bool) []string {
	var missing []string
	for reqKey, reqVal := range requiredTags {
		found := false
		for key, value := range effectiveTags {
			if caseInsensitive {
				if !strings.EqualFold(key, reqKey) {
					continue
				}
				if reqVal != "" && !strings.EqualFold(value, reqVal) {
					continue
				}
				found = true
				break
			} else {
				if key != reqKey {
					continue
				}
				if reqVal != "" && value != reqVal {
					continue
				}
				found = true
				break
			}
		}
		if !found {
			// If a value is specified, include it in the missing output.
			if reqVal != "" {
				missing = append(missing, fmt.Sprintf("%s:%s", reqKey, reqVal))
			} else {
				missing = append(missing, reqKey)
			}
		}
	}
	return missing
}
