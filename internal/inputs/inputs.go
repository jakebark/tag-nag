package inputs

import (
	"fmt"
	"os"
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
		fmt.Println("Error: Please specify a directory or file to scan.")
		os.Exit(1)
	}
	dir := pflag.Arg(0)

	if tags == "" {
		fmt.Println("Error: Please specify required tags using --tags")
		os.Exit(1)
	}

	requiredTags := cleanTagsInput(tags, ",")

	return UserInput{
		Directory:       dir,
		RequiredTags:    requiredTags,
		CaseInsensitive: caseInsensitive,
	}
}

func cleanTagsInput(s, sep string) []string {
	parts := strings.Split(s, sep)
	var trimmed []string
	for _, part := range parts {
		trimmed = append(trimmed, strings.TrimSpace(part))
	}
	return trimmed
}

