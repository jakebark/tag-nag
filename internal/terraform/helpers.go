package terraform

import (
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

func extractTraversalString(expr hcl.Expression, caseInsensitive bool) string {
	ste, ok := expr.(*hclsyntax.ScopeTraversalExpr)
	if !ok {
		return ""
	}
	tokens := []string{}
	for _, step := range ste.Traversal {
		switch t := step.(type) {
		case hcl.TraverseRoot:
			tokens = append(tokens, t.Name)
		case hcl.TraverseAttr:
			tokens = append(tokens, t.Name)
		}
	}
	result := strings.Join(tokens, ".")
	if caseInsensitive {
		result = strings.ToLower(result)
	}
	return result
}

func mergeTags(tagMaps ...map[string]bool) map[string]bool {
	merged := make(map[string]bool)
	for _, m := range tagMaps {
		for k, v := range m {
			merged[k] = v
		}
	}
	return merged
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
