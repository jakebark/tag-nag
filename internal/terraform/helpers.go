package terraform

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/jakebark/tag-nag/internal/shared"
	"github.com/zclconf/go-cty/cty"
)

// traversalToString converts a hcl hierachical/traversal string to a literal string
func traversalToString(expr hcl.Expression, caseInsensitive bool) string {
	if ste, ok := expr.(*hclsyntax.ScopeTraversalExpr); ok {
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
	// fallback - attempt to evaluate the expression as a literal value
	if v, diags := expr.Value(nil); !diags.HasErrors() {
		if v.Type().Equals(cty.String) {
			s := v.AsString()
			if caseInsensitive {
				s = strings.ToLower(s)
			}
			return s
		} else {
			return fmt.Sprintf("%v", v)
		}
	}
	return ""
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
		objExpr, ok := attr.Expr.(*hclsyntax.ObjectConsExpr)
		if !ok {
			return nil, fmt.Errorf("failed to extract tag map")
		}
		tags := make(TagMap)
		for _, item := range objExpr.Items {
			var keyStr string
			if v, vdiags := item.KeyExpr.Value(nil); !vdiags.HasErrors() {
				keyStr = v.AsString()
			} else {
				keyStr = traversalToString(item.KeyExpr, caseInsensitive)
			}

			var valLiteral string
			if v, vdiags := item.ValueExpr.Value(nil); !vdiags.HasErrors() {
				valLiteral = v.AsString()
			} else {
				valLiteral = traversalToString(item.ValueExpr, caseInsensitive)
			}
			tags[keyStr] = []string{valLiteral}
		}
		return tags, nil
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

// resolveTagValue recursively resolves a tag value with vars or locals
func resolveTagValue(value string, refMap TagReferences) string {
	// handles direct locals and vars
	if strings.HasPrefix(value, "local.") || strings.HasPrefix(value, "var.") {
		if tagMap, ok := refMap[value]; ok {
			if valList, found := tagMap["_"]; found && len(valList) > 0 {
				return valList[0]
			}
			for _, val := range tagMap {
				if len(val) > 0 {
					return val[0]
				}
			}
		}
		return value
	}

	// identify interpolation
	if !strings.Contains(value, "${") {
		return value
	}
	// match interpolation
	re := regexp.MustCompile(`\${([^}]+)}`)
	resolved := value

	// loop through interpolation(s)
	for {
		matches := re.FindAllStringSubmatch(resolved, -1)
		if len(matches) == 0 {
			break
		}
		for _, match := range matches {
			ref := match[1]
			replacement := ""
			// direct locals and vars
			if strings.HasPrefix(ref, "local.") || strings.HasPrefix(ref, "var.") {
				if tagMap, ok := refMap[ref]; ok {
					if valList, found := tagMap["_"]; found && len(valList) > 0 {
						replacement = valList[0]
					} else {
						for _, val := range tagMap {
							if len(val) > 0 {
								replacement = val[0]
								break
							}
						}
					}
				}
			} else {
				// indirect locals and vars
				if tagMap, ok := refMap["local."+ref]; ok {
					if valList, found := tagMap["_"]; found && len(valList) > 0 {
						replacement = valList[0]
					} else {
						for _, val := range tagMap {
							if len(val) > 0 {
								replacement = val[0]
								break
							}
						}
					}
				}
				if replacement == "" {
					if tagMap, ok := refMap["var."+ref]; ok {
						if valList, found := tagMap["_"]; found && len(valList) > 0 {
							replacement = valList[0]
						} else {
							for _, val := range tagMap {
								if len(val) > 0 {
									replacement = val[0]
									break
								}
							}
						}
					}
				}
			}
			resolved = strings.Replace(resolved, match[0], replacement, -1)
		}
	}
	return resolved
}

// SkipResource determines if a resource block should be skipped
func SkipResource(block *hclsyntax.Block, lines []string) bool {
	index := block.DefRange().Start.Line
	if index < len(lines) {
		if strings.Contains(lines[index], shared.TagNagIgnore) {
			return true
		}
	}
	return false
}
