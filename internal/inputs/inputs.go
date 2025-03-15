package inputs

import (
	"log"
	"strings"

	"github.com/spf13/pflag"
)

type TagMap map[string]string

type UserInput struct {
	Directory       string
	RequiredTags    TagMap
	CaseInsensitive bool
	ShowSkips       bool
}

// ParseFlags returns pased CLI flags and arguments
func ParseFlags() UserInput {
	var caseInsensitive bool
	var tags string
	var showSkips bool

	pflag.BoolVarP(&caseInsensitive, "case-insensitive", "c", false, "Make tag checks non-case-sensitive")
	pflag.BoolVarP(&showSkips, "skips", "s", false, "Show skips")
	pflag.StringVar(&tags, "tags", "", "Comma-separated list of required tag keys (e.g., 'Environment,Owner')")
	pflag.Parse()

	if pflag.NArg() < 1 {
		log.Fatal("Error: Please specify a directory or file to scan.")
	}
	if tags == "" {
		log.Fatal("Error: Please specify required tags using --tags")
	}

	return UserInput{
		Directory:       pflag.Arg(0),
		RequiredTags:    parseTags(tags),
		CaseInsensitive: caseInsensitive,
		ShowSkips:       showSkips,
	}
}

func parseTags(input string) TagMap {
	tagMap := make(map[string]string)
	pairs := strings.Split(input, ",")
	for _, pair := range pairs { // split on ,
		trimmed := strings.TrimSpace(pair)
		if trimmed == "" {
			continue
		}
		kv := strings.SplitN(trimmed, "=", 2) // split on =
		key := strings.TrimSpace(kv[0])
		var value string
		if len(kv) > 1 {
			value = strings.TrimSpace(kv[1])
		}
		tagMap[key] = value
	}
	return tagMap
}
