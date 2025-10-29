package inputs

import (
	"fmt"
	"log"
	"strings"

	"github.com/jakebark/tag-nag/internal/shared"
	"github.com/spf13/pflag"
)

type UserInput struct {
	Directory       string
	RequiredTags    shared.TagMap
	CaseInsensitive bool
	DryRun          bool
	CfnSpecPath     string
	Skip            []string
}

// ParseFlags returns pased CLI flags and arguments
func ParseFlags() UserInput {
	var caseInsensitive bool
	var dryRun bool
	var tags string
	var cfnSpecPath string
	var skip string

	pflag.BoolVarP(&caseInsensitive, "case-insensitive", "c", false, "Make tag checks non-case-sensitive")
	pflag.BoolVarP(&dryRun, "dry-run", "d", false, "Dry run tag:nag without triggering exit(1) code")
	pflag.StringVar(&tags, "tags", "", "Comma-separated list of required tag keys (e.g., 'Owner,Environment[Dev,Prod]')")
	pflag.StringVar(&cfnSpecPath, "cfn-spec", "", "Optional path to CloudFormationResourceSpecification.json)")
	pflag.StringVarP(&skip, "skip", "s", "", "Comma-separated list of files or directories to skip")
	pflag.Parse()

	if pflag.NArg() < 1 {
		log.Fatal("Error: Please specify a directory or file to scan")
	}

	// try config file if no tags provided
	if tags == "" {
		configFile, err := FindAndLoadConfigFile()
		if err != nil {
			log.Fatalf("Error loading config: %v", err)
		}
		if configFile != nil {
			return UserInput{
				Directory:       pflag.Arg(0),
				RequiredTags:    configFile.convertToTagMap(),
				CaseInsensitive: configFile.Settings.CaseInsensitive,
				DryRun:          configFile.Settings.DryRun,
				CfnSpecPath:     configFile.Settings.CfnSpec,
				Skip:            configFile.Skip,
			}
		}
		log.Fatal("Error: Please specify required tags using --tags or create .tag-nag.yml")
	}

	parsedTags, err := parseTags(tags)
	if err != nil {
		log.Fatalf("Error parsing tags: %v", err)
	}

	var skipPaths []string
	if skip != "" {
		skipPaths = strings.Split(skip, ",")
		for i := range skipPaths {
			skipPaths[i] = strings.TrimSpace(skipPaths[i])
		}
	}

	return UserInput{
		Directory:       pflag.Arg(0),
		RequiredTags:    parsedTags,
		CaseInsensitive: caseInsensitive,
		DryRun:          dryRun,
		CfnSpecPath:     cfnSpecPath,
		Skip:            skipPaths,
	}
}

// parses tag input components
func parseTags(input string) (shared.TagMap, error) {
	tagMap := make(shared.TagMap)
	pairs := splitTags(input)
	for _, pair := range pairs {
		trimmed := strings.TrimSpace(pair)
		if trimmed == "" {
			continue
		}

		key, values, err := parseTag(trimmed)
		if err != nil {
			return nil, fmt.Errorf("failed to parse tag component '%s': %w", trimmed, err)
		}
		tagMap[key] = values
	}
	return tagMap, nil
}

// parses tag keys and values
func parseTag(tagComponent string) (key string, values []string, err error) {
	trimmed := strings.TrimSpace(tagComponent)
	if trimmed == "" {
		return "", nil, fmt.Errorf("empty tag component")
	}

	// key and value
	openBracketIdx := strings.Index(trimmed, "[")
	if openBracketIdx != -1 {
		if !strings.HasSuffix(trimmed, "]") {
			return "", nil, fmt.Errorf("invalid tag format: %s. Expected closing ']'", trimmed)
		}

		key = strings.TrimSpace(trimmed[:openBracketIdx])
		if key == "" {
			return "", nil, fmt.Errorf("empty key in bracket format: %s", trimmed)
		}

		valuesStr := trimmed[openBracketIdx+1 : len(trimmed)-1]
		if valuesStr == "" {
			return key, []string{}, nil
		}

		valParts := strings.Split(valuesStr, ",")
		for _, v := range valParts {
			trimmedVal := strings.TrimSpace(v)
			if trimmedVal != "" {
				values = append(values, trimmedVal)
			}
		}
		return key, values, nil
	}

	// key only
	if strings.Contains(trimmed, "[") || strings.Contains(trimmed, "]") {
		return "", nil, fmt.Errorf("invalid tag format: %s. Contains '[' or ']' without matching pair or value definition", trimmed)
	}
	return trimmed, []string{}, nil
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
