package terraform

import (
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

type Violation struct {
	resourceType string
	resourceName string
	line         int
	missingTags  []string
}

func checkResources(body *hclsyntax.Body, requiredTags []string, defaultTagsByProvider map[string]map[string]bool, caseInsensitive bool) []Violation {
	var violations []Violation

	for _, block := range body.Blocks {
		if block.Type != "resource" || len(block.Labels) < 2 {
			continue
		}

		resourceType := block.Labels[0]
		resourceName := block.Labels[1]

		// only AWS resources
		if !strings.HasPrefix(resourceType, "aws_") {
			continue
		}

		// determine provider
		providerID := getResourceProvider(block, caseInsensitive)
		providerDefaults := defaultTagsByProvider[providerID]
		if providerDefaults == nil {
			providerDefaults = make(map[string]bool)
		}

		// resource tags
		resourceTags, _ := findTags(block, caseInsensitive)
		normalizedResourceTags := normalizeTagMap(resourceTags, caseInsensitive)

		// union of provider defaults and resource tags
		effectiveTags := make(map[string]bool)
		for k, v := range providerDefaults {
			effectiveTags[k] = v
		}
		for k, v := range normalizedResourceTags {
			effectiveTags[k] = v
		}

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

// checks for a "provider" attribute; if absent, returns "aws" as default.
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

		if ste, ok := attr.Expr.(*hclsyntax.ScopeTraversalExpr); ok {
			tokens := []string{}
			for _, step := range ste.Traversal {
				switch t := step.(type) {
				case hcl.TraverseRoot:
					tokens = append(tokens, t.Name)
				case hcl.TraverseAttr:
					tokens = append(tokens, t.Name)
				}
			}
			s := strings.Join(tokens, ".")
			if caseInsensitive {
				s = strings.ToLower(s)
			}
			return s
		}
	}

	defaultProvider := "aws"
	if caseInsensitive {
		defaultProvider = strings.ToLower(defaultProvider)
	}
	return defaultProvider
}
