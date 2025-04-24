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
func mergeTags(tagMaps ...shared.TagMap) shared.TagMap {
	merged := make(shared.TagMap)
	for _, m := range tagMaps {
		for k, v := range m {
			merged[k] = v
		}
	}
	return merged
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
