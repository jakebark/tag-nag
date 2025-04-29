package terraform

import (
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/jakebark/tag-nag/internal/config"
	"github.com/jakebark/tag-nag/internal/shared"
	"github.com/zclconf/go-cty/cty"
	ctyjson "github.com/zclconf/go-cty/cty/json"
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
		if strings.Contains(lines[index], config.TagNagIgnore) {
			return true
		}
	}
	return false
}

func convertCtyValueToString(val cty.Value) (string, error) {
	if !val.IsKnown() {
		return "", fmt.Errorf("value is unknown")
	}
	if val.IsNull() {
		return "", nil
	}

	ty := val.Type()
	switch {
	case ty == cty.String:
		return val.AsString(), nil
	case ty == cty.Number:
		bf := val.AsBigFloat()
		return bf.Text('f', -1), nil
	case ty == cty.Bool:
		return fmt.Sprintf("%t", val.True()), nil
	case ty.IsListType() || ty.IsTupleType() || ty.IsSetType() || ty.IsMapType() || ty.IsObjectType():

		simpleJSON, err := ctyjson.SimpleJSONValue{Value: val}.MarshalJSON()
		if err != nil {
			return "", fmt.Errorf("failed to marshal complex type to json: %w", err)
		}
		strJSON := string(simpleJSON)
		if len(strJSON) >= 2 && strJSON[0] == '"' && strJSON[len(strJSON)-1] == '"' {
			var unquotedStr string
			if err := json.Unmarshal(simpleJSON, &unquotedStr); err == nil {
				return unquotedStr, nil
			}
		}
		return strJSON, nil
	default:
		return fmt.Sprintf("%v", val), nil // Best effort
	}
}

// loadTaggableResources calls the Terraform JSON schema and returns a set of all resources that are taggable
func loadTaggableResources(providerAddr string) map[string]bool {
	out, err := exec.Command(
		"terraform", "providers", "schema", "-json",
	).Output()
	if err != nil {
		log.Printf("failed to load AWS terraform provider schema: %v", err)
		return nil
	}

	// unmarshall what we need
	var s struct {
		ProviderSchemas map[string]struct {
			ResourceSchemas map[string]struct {
				Block struct {
					Attributes map[string]json.RawMessage `json:"attributes"`
				} `json:"block"`
			} `json:"resource_schemas"`
		} `json:"provider_schemas"`
	}
	if err := json.Unmarshal(out, &s); err != nil {
		log.Printf("failed to parse schema JSON: %v", err)
		return nil
	}

	taggable := make(map[string]bool)
	if ps, ok := s.ProviderSchemas[providerAddr]; ok {
		for resType, schema := range ps.ResourceSchemas {
			if _, has := schema.Block.Attributes["tags"]; has {
				taggable[resType] = true
			}
		}
	}
	return taggable
}
