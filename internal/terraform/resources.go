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

		// only AWS resources
		if !strings.HasPrefix(resourceType, "aws_") {
			continue
		}

		// default tags magic
		providerID := getResourceProvider(block, caseInsensitive)
		providerDefaults := defaultTags.ProviderTags[providerID]
		if providerDefaults == nil {
			providerDefaults = make(map[string]bool)
		}

		// resource tags
		resourceTags, err := findTags(block, caseInsensitive)
		if err != nil {
			resourceTags = make(map[string]bool)
		}
		effectiveTags := mergeTags(providerDefaults, resourceTags)

		// determine which tags are missing
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

		val, diags := attr.Expr.Value(nil)
		if !diags.HasErrors() {
			s := val.AsString()
			if caseInsensitive {
				s = strings.ToLower(s)
			}
			return s
		}

		s := extractTraversalString(attr.Expr, caseInsensitive)
		if s != "" {
			return s
		}

	}

	defaultProvider := "aws"
	if caseInsensitive {
		defaultProvider = strings.ToLower(defaultProvider)
	}
	return defaultProvider
}
