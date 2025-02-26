package terraform

import (
	"strings"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

func findTags(block *hclsyntax.Block, caseInsensitive bool) (map[string]bool, error) {
	if attr, ok := block.Body.Attributes["tags"]; ok {
		value, diags := attr.Expr.Value(nil)
		if diags.HasErrors() {
			return nil, diags.Errs()[0]
		}
		if value.Type().IsObjectType() {
			return extractTagKeys(value.AsValueMap(), caseInsensitive), nil
		}
	}
	return make(map[string]bool), nil
}

func extractTagKeys(tagMap map[string]cty.Value, caseInsensitive bool) map[string]bool {
	tags := make(map[string]bool)
	for key := range tagMap {
		if caseInsensitive {
			key = strings.ToLower(key)
		}
		tags[key] = true
	}
	return tags
}

func normalizeTagMap(tagMap map[string]bool, caseInsensitive bool) map[string]bool {
	if !caseInsensitive {
		return tagMap
	}
	normalized := make(map[string]bool)
	for key := range tagMap {
		normalized[strings.ToLower(key)] = true
	}
	return normalized
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
