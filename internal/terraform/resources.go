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

func checkResources(body *hclsyntax.Body, requiredTags []string, defaultTags *DefaultTags, caseInsensitive bool) []Violation {
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
