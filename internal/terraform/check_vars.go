package terraform

import (
	"os"
	"path/filepath"
	"strings"
)

func checkVariables(dirPath string) map[string]map[string]bool {
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
