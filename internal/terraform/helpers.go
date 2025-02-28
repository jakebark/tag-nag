package terraform

import (
	"fmt"
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

	// converts hcl qualified name into a literal string
	// eg local.tags to ... local.tags
	// hcl sees it as hierachical components, eg:
	// TraversalExpr{
	// Traversal: [
	//    TraverseRoot{Name: "local"},
	//    TraverseAttr{Name: "tags"}
	// ]}
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

func extractTags(attr *hclsyntax.Attribute, caseInsensitive bool) map[string]bool {
	tags, err := extractTagMap(attr, caseInsensitive)
	if err != nil {
		// todo error logging
		return make(map[string]bool)
	}
	return tags
}

func extractTagMap(attr *hclsyntax.Attribute, caseInsensitive bool) (map[string]bool, error) {
	val, diags := attr.Expr.Value(nil)
	// check for errors in tags
	if diags.HasErrors() || !val.Type().IsObjectType() {
		return nil, fmt.Errorf("failed to extract tag map")
	}

	tags := make(map[string]bool)
	// convert hcl to go map
	for key := range val.AsValueMap() {
		if caseInsensitive {
			key = strings.ToLower(key)
		}
		tags[key] = true
	}
	return tags, nil
}
