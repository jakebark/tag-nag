package terraform

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/jakebark/tag-nag/internal/shared"
)

// traversalToString converts a hcl hierachical/traversal string to a literal string.
func traversalToString(expr hcl.Expression, caseInsensitive bool) string {
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

// mergeTags combines multiple tag maps
func mergeTags(tagMaps ...TagMap) TagMap {
	merged := make(TagMap)
	for _, m := range tagMaps {
		for k, v := range m {
			merged[k] = v
		}
	}
	return merged
}

// extractTags handles errors in extracting tags from hcl
func extractTags(attr *hclsyntax.Attribute, caseInsensitive bool) TagMap {
	tags, err := extractTagMap(attr, caseInsensitive)
	if err != nil {
		// todo error logging
		return make(TagMap)
	}
	return tags
}

// extractTagMap extracts the hcl tag map to a go map
func extractTagMap(attr *hclsyntax.Attribute, caseInsensitive bool) (TagMap, error) {
	val, diags := attr.Expr.Value(nil)

	if diags.HasErrors() || !val.Type().IsObjectType() {
		return nil, fmt.Errorf("failed to extract tag map")
	}

	tags := make(TagMap)
	for key, value := range val.AsValueMap() {
		if caseInsensitive {
			key = strings.ToLower(key)
		}
		tags[key] = []string{value.AsString()}
	}
	return tags, nil
}

func skipResource(block *hclsyntax.Block, lines []string) bool {
	index := block.DefRange().Start.Line
	if index < len(lines) {
		if strings.Contains(lines[index], shared.TagNagIgnore) {
			return true
		}
	}
	return false
}

// resolveTagValue recursively resolves a tag value with vars or locals
func resolveTagValue(value string, refMap TagReferences) string {
	// If the value does not contain any interpolation or reference patterns, return as is.
	if !strings.Contains(value, "${") && !strings.Contains(value, "local.") && !strings.Contains(value, "var.") {
		return value
	}

	// Handle interpolation, e.g. "${var.env}-env-${local.account}"
	re := regexp.MustCompile(`\${([^}]+)}`)
	resolved := value

	// Loop until no more interpolations are found.
	for {
		matches := re.FindAllStringSubmatch(resolved, -1)
		if len(matches) == 0 {
			break
		}
		for _, match := range matches {
			ref := match[1]
			replacement := ""

			// If the reference already starts with "local." or "var.", do a direct lookup.
			if strings.HasPrefix(ref, "local.") || strings.HasPrefix(ref, "var.") {
				if tagMap, ok := refMap[ref]; ok {
					for _, val := range tagMap {
						if len(val) > 0 {
							replacement = val[0]
						}
						break
					}
				}
			} else {
				// Try looking up with the "local." prefix.
				if tagMap, ok := refMap["local."+ref]; ok {
					for _, val := range tagMap {
						if len(val) > 0 {
							replacement = val[0]
						}
						break
					}
				}
				// If not found, try "var." prefix.
				if replacement == "" {
					if tagMap, ok := refMap["var."+ref]; ok {
						for _, val := range tagMap {
							if len(val) > 0 {
								replacement = val[0]
							}
							break
						}
					}
				}
			}

			// If replacement not found, leave it as empty (which effectively “ignores” missing references).
			resolved = strings.Replace(resolved, match[0], replacement, -1)
		}
	}
	return resolved
}
