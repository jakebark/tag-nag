package terraform

import (
	"strings"

	"github.com/hashicorp/hcl/v2/hclsyntax"
)

type Violation struct {
	resourceType string
	resourceName string
	line         int
	missingTags  []string
}

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

		resourceTags := findTags(block, caseInsensitive)

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

func findTags(block *hclsyntax.Block, caseInsensitive bool) map[string]bool {
	if attr, ok := block.Body.Attributes["tags"]; ok {
		return extractTags(attr, caseInsensitive)
	}
	return make(map[string]bool)
}

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
