package inputs

import (
	"log"
	"strings"

	"github.com/spf13/pflag"
)

type UserInput struct {
	Directory       string
	RequiredTags    []string
	CaseInsensitive bool
}

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
		RequiredTags:    cleanTags(tags),
		CaseInsensitive: caseInsensitive,
	}
}

func cleanTags(input string) []string {
	tags := strings.Split(input, ",")
	for i := range tags {
		tags[i] = strings.TrimSpace(tags[i])
	}
	return tags
}
