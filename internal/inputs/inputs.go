package inputs

import (
	"log"
	"strings"

	"github.com/spf13/pflag"
)

// UserInput holds the user CLI inputs
type UserInput struct {
	Directory       string
	RequiredTags    []string
	CaseInsensitive bool
}

// ParseFlags parses CLI flags and arguments
// It returns a struct containing the parsed values
func ParseFlags() UserInput {
	var caseInsensitive bool
	var tags string

	pflag.BoolVarP(&caseInsensitive, "case-insensitive", "c", false, "Make tag checks non-case-sensitive")
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
		RequiredTags:    sliceAndTrim(tags),
		CaseInsensitive: caseInsensitive,
	}
}

// sliceAndTrim removes whitespace and converts the CLI string input to a slice
func sliceAndTrim(input string) []string {
	tags := strings.Split(input, ",")
	for i := range tags {
		tags[i] = strings.TrimSpace(tags[i])
	}
	return tags
}
