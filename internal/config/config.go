package config

import (
	"github.com/zclconf/go-cty/cty/function"
	"github.com/zclconf/go-cty/cty/function/stdlib"
)

const (
	TagNagIgnore    = "#tag-nag ignore"
	TagNagIgnoreAll = "#tag-nag ignore-all"
)

var StdlibFuncs = map[string]function.Function{
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
