package inputs

import (
	"log"
	"strings"

	"github.com/spf13/pflag"
)

type TagMap map[string][]string

type UserInput struct {
	Directory       string
	RequiredTags    TagMap
	CaseInsensitive bool
	DryRun          bool
}

// ParseFlags returns pased CLI flags and arguments
func ParseFlags() UserInput {
	var caseInsensitive bool
	var dryRun bool
	var tags string

	pflag.BoolVarP(&caseInsensitive, "case-insensitive", "c", false, "Make tag checks non-case-sensitive")
	pflag.BoolVarP(&dryRun, "dry-run", "d", false, "Dry run tag:nag without triggering exit(1) code")
	pflag.StringVar(&tags, "tags", "", "Comma-separated list of required tag keys (e.g., 'Owner,Environment[Dev,Prod]')")
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
		DryRun:          dryRun,
	}
}

func parseTags(input string) TagMap {
	tagMap := make(TagMap)
	pairs := splitTags(input)
	for _, pair := range pairs { //split on ,
		trimmed := strings.TrimSpace(pair)
		if trimmed == "" {
			continue
		}

		// legacy "="
		// rm in later version
		if eqIdx := strings.Index(trimmed, "="); eqIdx != -1 {
			key := strings.TrimSpace(trimmed[:eqIdx])
			value := strings.TrimSpace(trimmed[eqIdx+1:])
			tagMap[key] = []string{value}
			continue
		}

		// is [ present
		if openIdx := strings.Index(trimmed, "["); openIdx != -1 {
			// check for closing
			if trimmed[len(trimmed)-1] != ']' {
				log.Fatalf("Invalid tag format: %s. Expected closing ']'", trimmed)
			}
			// get key
			key := strings.TrimSpace(trimmed[:openIdx])
			// get value
			valuesStr := trimmed[openIdx+1 : len(trimmed)-1]
			values := []string{}
			// get values
			for _, v := range strings.Split(valuesStr, ",") {
				v = strings.TrimSpace(v)
				if v != "" {
					values = append(values, v)
				}
			}
			tagMap[key] = values
		} else {
			// just tag key
			tagMap[trimmed] = []string{}
		}
	}
	return tagMap
}

// splitTags splits the input string on commas outside of brackets
// to fix the [a,b,c] issue
func splitTags(input string) []string {
	var parts []string
	start := 0
	depth := 0
	for i, r := range input {
		switch r {
		case '[':
			depth++
		case ']':
			if depth > 0 {
				depth--
			}
		case ',':
			if depth == 0 {
				parts = append(parts, strings.TrimSpace(input[start:i]))
				start = i + 1
			}
		}
	}
	parts = append(parts, strings.TrimSpace(input[start:]))
	return parts
}
