package config

import (
	"github.com/zclconf/go-cty/cty/function"
	"github.com/zclconf/go-cty/cty/function/stdlib"
)

const (
	TagNagIgnore    = "#tag-nag ignore"
	TagNagIgnoreAll = "#tag-nag ignore-all"
)

var SkippedDirs = []string{
	".terraform",
	".git",
}

// terraform functions, used when evaluating context of locals and vars
// added manually, no reasonable workaround to auto-import all
// https://developer.hashicorp.com/terraform/language/functions
// https://pkg.go.dev/github.com/zclconf/go-cty@v1.16.2/cty/function/stdlib
var StdlibFuncs = map[string]function.Function{
	"chomp":      stdlib.ChompFunc,
	"coalesce":   stdlib.CoalesceFunc,
	"concat":     stdlib.ConcatFunc,
	"flatten":    stdlib.FlattenFunc,
	"join":       stdlib.JoinFunc,
	"jsondecode": stdlib.JSONDecodeFunc,
	"keys":       stdlib.KeysFunc,
	"lower":      stdlib.LowerFunc,
	"merge":      stdlib.MergeFunc,
	"min":        stdlib.MinFunc,
	"max":        stdlib.MaxFunc,
	"regex":      stdlib.RegexFunc,
	"regexall":   stdlib.RegexAllFunc,
	"setunion":   stdlib.SetUnionFunc,
	"slice":      stdlib.SliceFunc,
	"split":      stdlib.SplitFunc,
	"sort":       stdlib.SortFunc,
	"trim":       stdlib.TrimFunc,
	"trimprefix": stdlib.TrimPrefixFunc,
	"trimspace":  stdlib.TrimSpaceFunc,
	"trimsuffix": stdlib.TrimSuffixFunc,
	"upper":      stdlib.UpperFunc,
	"values":     stdlib.ValuesFunc,
}
