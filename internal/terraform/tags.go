package terraform

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

func findTags(block *hclsyntax.Block, caseInsensitive bool) (map[string]bool, error) {
	tags := make(map[string]bool)
	if attr, ok := block.Body.Attributes["tags"]; ok {
		expr := attr.Expr
		value, diags := expr.Value(nil)
		if diags.HasErrors() {
			return nil, diags.Errs()[0]
		}
		if value.Type().IsObjectType() {
			for key := range value.AsValueMap() {
				if caseInsensitive {
					key = strings.ToLower(key)
				}
				tags[key] = true
			}
		}
	}
	return tags, nil
}

func findDefaultTagsByProvider(dirPath string, caseInsensitive bool) map[string]map[string]bool {
	result := make(map[string]map[string]bool)
	filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".tf") {
			parser := hclparse.NewParser()
			file, diags := parser.ParseHCLFile(path)
			if diags.HasErrors() {
				return nil
			}
			syntaxBody, ok := file.Body.(*hclsyntax.Body)
			if !ok {
				return nil
			}
			for _, block := range syntaxBody.Blocks {
				if block.Type == "provider" && len(block.Labels) > 0 {
					providerName := block.Labels[0]
					alias := ""
					if attr, ok := block.Body.Attributes["alias"]; ok {
						val, diags := attr.Expr.Value(nil)
						if !diags.HasErrors() {
							alias = val.AsString()
						}
					}
					var providerID string
					if alias != "" {
						providerID = providerName + "." + alias
					} else {
						providerID = providerName
					}
					if caseInsensitive {
						providerID = strings.ToLower(providerID)
					}
					// Look for a default_tags block inside this provider block.
					for _, subBlock := range block.Body.Blocks {
						if subBlock.Type == "default_tags" {
							tags, _ := findTags(subBlock, caseInsensitive)
							result[providerID] = tags
						}
					}
				}
			}
		}
		return nil
	})
	if len(result) > 0 {
		fmt.Printf("🔍 Found default tags: %v\n", result)
	} else {
		fmt.Println("⚠️  No default tags found")
	}
	return result
}

// return a new map with lowercased keys if caseInsensitive is enabled.
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

// return the required tags (with original case) not found in effectiveTags
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
