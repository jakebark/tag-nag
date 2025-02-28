package terraform

import (
	"fmt"
	"strings"

	"github.com/hashicorp/hcl/v2/hclsyntax"
)

// single tag extraction func
func getTagMap(attr *hclsyntax.Attribute, caseInsensitive bool) (map[string]bool, error) {
	val, diags := attr.Expr.Value(nil)
	if diags.HasErrors() || !val.Type().IsObjectType() {
		return nil, fmt.Errorf("failed to extract tag map")
	}
	tags := make(map[string]bool)
	for key := range val.AsValueMap() {
		if caseInsensitive {
			key = strings.ToLower(key)
		}
		tags[key] = true
	}
	return tags, nil
}

func findTags(block *hclsyntax.Block, caseInsensitive bool) (map[string]bool, error) {
	if attr, ok := block.Body.Attributes["tags"]; ok {
		return getTagMap(attr, caseInsensitive)
	}
	return make(map[string]bool), nil
}
