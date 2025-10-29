package config

import (
	"github.com/zclconf/go-cty/cty/function"
	"github.com/zclconf/go-cty/cty/function/stdlib"
)

const (
	TagNagIgnore       = "#tag-nag ignore"
	TagNagIgnoreAll    = "#tag-nag ignore-all"
	DefaultConfigFile  = ".tag-nag.yml"
	AltConfigFile      = ".tag-nag.yaml"
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
	"abs":          stdlib.AbsoluteFunc,
	"ceil":         stdlib.CeilFunc,
	"chomp":        stdlib.ChompFunc,
	"chunklist":    stdlib.ChunklistFunc,
	"coalesce":     stdlib.CoalesceFunc,
	"coalescelist": stdlib.CoalesceListFunc,
	"compact":      stdlib.CompactFunc,
	"concat":       stdlib.ConcatFunc,
	"contains":     stdlib.ContainsFunc,
	"csvdecode":    stdlib.CSVDecodeFunc,
	"distinct":     stdlib.DistinctFunc,
	"element":      stdlib.ElementFunc,
	"flatten":      stdlib.FlattenFunc,
	"floor":        stdlib.FloorFunc,
	"format":       stdlib.FormatFunc,
	"formatdate":   stdlib.FormatDateFunc,
	"formatlist":   stdlib.FormatListFunc,
	"indent":       stdlib.IndentFunc,
	"index":        stdlib.IndexFunc,
	"int":          stdlib.IntFunc,
	"join":         stdlib.JoinFunc,
	"jsondecode":   stdlib.JSONDecodeFunc,
	"jsonencode":   stdlib.JSONEncodeFunc,
	"keys":         stdlib.KeysFunc,
	"length":       stdlib.LengthFunc,
	"log":          stdlib.LogFunc,
	"lookup":       stdlib.LookupFunc,
	"lower":        stdlib.LowerFunc,
	"max":          stdlib.MaxFunc,
	"merge":        stdlib.MergeFunc,
	"min":          stdlib.MinFunc,
	"parseint":     stdlib.ParseIntFunc,
	"pow":          stdlib.PowFunc,
	"range":        stdlib.RangeFunc,
	"regex":        stdlib.RegexFunc,
	"regexall":     stdlib.RegexAllFunc,
	"regexreplace": stdlib.RegexReplaceFunc,
	"replace":      stdlib.ReplaceFunc,
	"reverse":      stdlib.ReverseFunc,
	"reverselist":  stdlib.ReverseListFunc,
	"setunion":     stdlib.SetUnionFunc,
	"slice":        stdlib.SliceFunc,
	"sort":         stdlib.SortFunc,
	"split":        stdlib.SplitFunc,
	"trim":         stdlib.TrimFunc,
	"trimprefix":   stdlib.TrimPrefixFunc,
	"trimspace":    stdlib.TrimSpaceFunc,
	"trimsuffix":   stdlib.TrimSuffixFunc,
	"upper":        stdlib.UpperFunc,
	"values":       stdlib.ValuesFunc,
}
