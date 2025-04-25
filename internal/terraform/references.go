package terraform

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
	"github.com/zclconf/go-cty/cty/function/stdlib"
)

var stdlibFuncs = map[string]function.Function{
	"upper":      stdlib.UpperFunc,
	"lower":      stdlib.LowerFunc,
	"chomp":      stdlib.ChompFunc,
	"coalesce":   stdlib.CoalesceFunc,
	"concat":     stdlib.ConcatFunc,
	"flatten":    stdlib.FlattenFunc,
	"merge":      stdlib.MergeFunc,
	"min":        stdlib.MinFunc,
	"max":        stdlib.MaxFunc,
	"regex":      stdlib.RegexFunc,
	"slice":      stdlib.SliceFunc,
	"trim":       stdlib.TrimFunc,
	"trimprefix": stdlib.TrimPrefixFunc,
	"trimspace":  stdlib.TrimSpaceFunc,
	"trimsuffix": stdlib.TrimSuffixFunc,
}

func buildTagContext(dirPath string) (*TerraformContext, error) {
	parsedFiles := make(map[string]*hcl.File)
	parser := hclparse.NewParser()

	// first pass, parse files
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() && (info.Name() == ".terraform" || info.Name() == ".git") {
			return filepath.SkipDir
		}
		if !info.IsDir() && filepath.Ext(path) == ".tf" {
			file, diags := parser.ParseHCLFile(path)
			if diags.HasErrors() {
				log.Printf("Error parsing HCL file %s: %v\n", path, diags)
			}
			if file != nil {
				parsedFiles[path] = file
			}
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("error walking directory %s: %w", dirPath, err)
	}

	if len(parsedFiles) == 0 {
		log.Println("No Terraform files (.tf) found to build context.")
		return &TerraformContext{
			EvalContext: &hcl.EvalContext{
				Variables: make(map[string]cty.Value),
				Functions: make(map[string]function.Function),
			},
		}, nil
	}

	// second pass, evaluate vars
	tfVars := make(map[string]cty.Value)
	for _, file := range parsedFiles {
		body, ok := file.Body.(*hclsyntax.Body)
		if !ok {
			continue
		}

		for _, block := range body.Blocks {
			if block.Type == "variable" && len(block.Labels) > 0 {
				varName := block.Labels[0]
				if defaultAttr, exists := block.Body.Attributes["default"]; exists {
					val, diags := defaultAttr.Expr.Value(nil)
					if diags.HasErrors() {
						log.Printf("Error evaluating default for variable %q: %v", varName, diags)
						val = cty.NullVal(cty.DynamicPseudoType)
					}
					tfVars[varName] = val
				} else {
					tfVars[varName] = cty.NullVal(cty.DynamicPseudoType)
				}
			}
		}
	}

	// 3rd pass, evaluate locals
	tfLocals := make(map[string]cty.Value)
	localsDefs := make(map[string]hcl.Expression)

	for _, file := range parsedFiles {
		body, ok := file.Body.(*hclsyntax.Body)
		if !ok {
			continue
		}
		for _, block := range body.Blocks {
			if block.Type == "locals" {
				for name, attr := range block.Body.Attributes {
					localsDefs[name] = attr.Expr
				}
			}
		}
	}

	evalCtxForLocals := &hcl.EvalContext{
		Variables: map[string]cty.Value{"var": cty.ObjectVal(tfVars)},
		Functions: stdlibFuncs,
	}
	evalCtxForLocals.Variables["local"] = cty.NullVal(cty.DynamicPseudoType) // Placeholder for local

	const maxLocalPasses = 10
	evaluatedCount := 0
	for pass := 0; pass < maxLocalPasses && evaluatedCount < len(localsDefs); pass++ {
		madeProgress := false

		evalCtxForLocals.Variables["local"] = cty.ObjectVal(tfLocals)

		for name, expr := range localsDefs {
			if _, exists := tfLocals[name]; exists {
				continue
			}

			val, diags := expr.Value(evalCtxForLocals)
			if !diags.HasErrors() {
				tfLocals[name] = val
				evaluatedCount++
				madeProgress = true
			}

		}
		if !madeProgress && evaluatedCount < len(localsDefs) {
			log.Printf("Warning: Could not resolve all locals dependencies after %d passes.", pass+1)

			break
		}
	}

	finalCtx := &hcl.EvalContext{
		Variables: map[string]cty.Value{
			"var":   cty.ObjectVal(tfVars),
			"local": cty.ObjectVal(tfLocals),
		},
		Functions: stdlibFuncs,
	}

	return &TerraformContext{EvalContext: finalCtx}, nil
}
