package terraform

import (
	"strings"

	"github.com/hashicorp/hcl/v2/hclsyntax"
)

// Violation is a non-compliant tag
type Violation struct {
	resourceType string
	resourceName string
	line         int
	missingTags  []string
}

// checkResourcesForTags inspects resource blocks and returns violations
func checkResourcesForTags(body *hclsyntax.Body, requiredTags []string, defaultTags *DefaultTags, caseInsensitive bool) []Violation {
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
		providerDefaultTags := defaultTags.ProviderTags[providerID]
		if providerDefaultTags == nil {
			providerDefaultTags = make(map[string]bool)
		}

		resourceTags := findTags(block, defaultTags.ReferencedTags, caseInsensitive)

		effectiveTags := mergeTags(providerDefaultTags, resourceTags)

		missingTags := filterMissingTags(requiredTags, effectiveTags, caseInsensitive)
		if len(missingTags) > 0 {
			violations = append(violations, Violation{
				resourceType: resourceType,
				resourceName: resourceName,
				line:         block.DefRange().Start.Line,
				missingTags:  missingTags,
			})
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
		s := extractTraversalString(attr.Expr, caseInsensitive)
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
func findTags(block *hclsyntax.Block, referencedTags map[string]map[string]bool, caseInsensitive bool) map[string]bool {
	if attr, ok := block.Body.Attributes["tags"]; ok {
		// literal tags
		tags := extractTags(attr, caseInsensitive)
		if len(tags) > 0 {
			return tags
		}

		// referenced tags
		refKey := extractTraversalString(attr.Expr, caseInsensitive)
		if refKey != "" {
			if resolved, ok := referencedTags[refKey]; ok {
				return resolved
			}
		}
	}
	return make(map[string]bool)
}

// filterMissingTags compares the literal tags against the required tags
func filterMissingTags(requiredTags []string, effectiveTags map[string]bool, caseInsensitive bool) []string {
	missing := []string{}
	for _, tag := range requiredTags {
		found := false
		for existing := range effectiveTags {
			if caseInsensitive {
				if strings.EqualFold(existing, tag) {
					found = true
					break
				}
			} else {
				if existing == tag {
					found = true
					break
				}
			}
		}
		if !found {
			missing = append(missing, tag)
		}
	}
	return missing
}
