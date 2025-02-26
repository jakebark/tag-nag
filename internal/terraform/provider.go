package terraform

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

func checkForDefaultTags(dirPath string, caseInsensitive bool) map[string]map[string]bool {
	result := make(map[string]map[string]bool)
	localsAndVars := extractLocalsAndVariables(dirPath)

	_ = filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() || !strings.HasSuffix(path, ".tf") {
			return nil
		}

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
				providerID := getProviderID(block, caseInsensitive)

				tags := extractDefaultTagsBlock(block, localsAndVars, caseInsensitive)

				if tags != nil {
					result[providerID] = tags
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

func getProviderID(block *hclsyntax.Block, caseInsensitive bool) string {
	providerName := block.Labels[0]
	alias := ""

	if attr, ok := block.Body.Attributes["alias"]; ok {
		val, diags := attr.Expr.Value(nil)
		if !diags.HasErrors() {
			alias = val.AsString()
		}
	}

	providerID := providerName
	if alias != "" {
		providerID += "." + alias
	}

	if caseInsensitive {
		providerID = strings.ToLower(providerID)
	}
	return providerID
}

func extractDefaultTagsBlock(block *hclsyntax.Block, localsAndVars map[string]map[string]bool, caseInsensitive bool) map[string]bool {
	for _, subBlock := range block.Body.Blocks {
		if subBlock.Type == "default_tags" {
			tags, _ := findTags(subBlock, caseInsensitive)

			if attr, exists := subBlock.Body.Attributes["tags"]; exists {
				tags = resolveTagReferences(attr, localsAndVars)
			}
			return tags
		}
	}
	return nil
}

func resolveTagReferences(attr *hclsyntax.Attribute, localsAndVars map[string]map[string]bool) map[string]bool {
	if traversalExpr, ok := attr.Expr.(*hclsyntax.ScopeTraversalExpr); ok {
		traversedParts := []string{}
		for _, step := range traversalExpr.Traversal {
			switch t := step.(type) {
			case hcl.TraverseRoot:
				traversedParts = append(traversedParts, t.Name)
			case hcl.TraverseAttr:
				traversedParts = append(traversedParts, t.Name)
			}
		}
		varRef := strings.Join(traversedParts, ".")

		if resolvedTags, found := localsAndVars[varRef]; found {
			return resolvedTags
		}
	}
	return nil
}

func extractLocalsAndVariables(dirPath string) map[string]map[string]bool {
	localsAndVars := make(map[string]map[string]bool)

	_ = filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
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
				if block.Type == "locals" {
					for name, attr := range block.Body.Attributes {
						val, diags := attr.Expr.Value(nil)
						if !diags.HasErrors() && val.Type().IsObjectType() {
							localTags := make(map[string]bool)
							for key := range val.AsValueMap() {
								localTags[key] = true
							}
							localsAndVars["local."+name] = localTags
						}
					}
				}

				if block.Type == "variable" && len(block.Labels) > 0 {
					varName := block.Labels[0]
					if attr, ok := block.Body.Attributes["default"]; ok {
						val, diags := attr.Expr.Value(nil)
						if !diags.HasErrors() && val.Type().IsObjectType() {
							varTags := make(map[string]bool)
							for key := range val.AsValueMap() {
								varTags[key] = true
							}
							localsAndVars["var."+varName] = varTags
						}
					}
				}
			}
		}
		return nil
	})

	return localsAndVars
}
