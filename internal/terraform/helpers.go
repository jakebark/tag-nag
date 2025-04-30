package terraform

import (
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"sort"
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
	log.Println("DEBUG: load provider schema")
	out, err := exec.Command(
		"terraform", "providers", "schema", "-json",
	).Output()
	if err != nil {
		log.Fatalf("failed to load AWS terraform provider schema: %v", err)
		return nil
	}
	log.Println("DEBUG: Successfully executed 'terraform providers schema -json'.")

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
		log.Fatalf("failed to parse schema JSON: %v", err)
		return nil
	}
	log.Println("DEBUG: Successfully parsed schema JSON.")

	taggable := make(map[string]bool)
	if ps, ok := s.ProviderSchemas[providerAddr]; ok {
		log.Printf("DEBUG: Found provider schema for: %s", providerAddr)
		for resType, schema := range ps.ResourceSchemas {
			if _, has := schema.Block.Attributes["tags"]; has {
				taggable[resType] = true
			} else { //
				taggable[resType] = false
			}
		}
	} else {
		log.Printf("DEBUG: Provider schema not found for: %s", providerAddr)
		return nil
	}
	log.Printf("Generated taggable map (Size: %d):", len(taggable))
	keys := make([]string, 0, len(taggable))
	for k := range taggable {
		keys = append(keys, k)
	}
	sort.Strings(keys) // Sort for consistent output
	// Limit printing if the map is very large
	limit := 200 // Print details for up to 200 resource types
	for i, k := range keys {
		if i < limit || k == "aws_kms_alias" || k == "aws_kms_key" { // Always print kms_alias/key if present
			log.Printf("  - %s: %t", k, taggable[k])
		} else if i == limit {
			log.Printf("  ... (output truncated)")
			break
		}
	}
	if isKmsAliasTaggable, ok := taggable["aws_kms_alias"]; ok {
		log.Printf("DEBUG: 'aws_kms_alias' found in taggable map, value: %t", isKmsAliasTaggable)
	} else {
		log.Printf("DEBUG: 'aws_kms_alias' NOT found in taggable map.")
	}

	return taggable
}
