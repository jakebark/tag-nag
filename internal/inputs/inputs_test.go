package inputs

import (
	"os"
	"os/exec"
	"reflect"
	"testing"

	"github.com/jakebark/tag-nag/internal/shared"
)

func testParseTags(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected shared.TagMap
	}{
		{
			name:  "key",
			input: "Owner",
			expected: shared.TagMap{
				"Owner": {},
			},
		},
		{
			name:  "multiple keys",
			input: "Owner, Environment , Project",
			expected: shared.TagMap{
				"Owner":       {},
				"Environment": {},
				"Project":     {},
			},
		},
		{
			name:  "mixed keys and values",
			input: "Owner[jake], Environment[Dev,Prod], CostCenter",
			expected: shared.TagMap{
				"Owner":       {"jake"},
				"Environment": {"Dev", "Prod"},
				"CostCenter":  {},
			},
		},
		{
			name:  "legacy value input",
			input: "Env=dev, Owner[jake]",
			expected: shared.TagMap{
				"Env":   {"dev"},
				"Owner": {"jake"},
			},
		},
		{
			name:     "empty",
			input:    "",
			expected: shared.TagMap{},
		},
		{
			name:     "whitespace",
			input:    "  ,   ",
			expected: shared.TagMap{},
		},
		{
			name:  "mixed keys and values, with whitespace",
			input: " Owner ,  Environment[Dev, Prod] ",
			expected: shared.TagMap{
				"Owner":       {},
				"Environment": {"Dev", "Prod"},
			},
		},
		{
			name:  "leading comma",
			input: ",Owner",
			expected: shared.TagMap{
				"Owner": {},
			},
		},
		{
			name:  "missing value",
			input: "Env[]",
			expected: shared.TagMap{
				"Env": {}, // No values extracted
			},
		},
		{
			name:  "missing value, other values present",
			input: "Env[Dev,,Prod]",
			expected: shared.TagMap{
				"Env": {"Dev", "Prod"},
			},
		},
		{
			name:  "whitespace preserved",
			input: "Owner[it belongs to me]",
			expected: shared.TagMap{
				"Owner": {"it belongs to me"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := parseTags(tc.input)
			if !reflect.DeepEqual(actual, tc.expected) {
				t.Errorf("parseTags(%q) = %v; want %v", tc.input, actual, tc.expected)
			}
		})
	}
}

func testSplitTags(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected []string
	}{
		{"empty", "", []string{""}},
		{"key", "Owner", []string{"Owner"}},
		{"multiple keys", "Owner, Env , Project", []string{"Owner", "Env", "Project"}},
		{"value", "Owner[Jake]", []string{"Owner[Jake]"}},
		{"multiple values", "Env[Dev,Prod]", []string{"Env[Dev,Prod]"}},
		{"mixed keys and values", "Owner[Jake], Env[Dev,Prod], CostCenter", []string{"Owner[Jake]", "Env[Dev,Prod]", "CostCenter"}},
		{"legacy value input", "owner=jake, env=prod", []string{"owner=jake", "env=prod"}},
		{"mixed legacy", "owner=jake, Env[Dev,Prod]", []string{"owner=jake", "Env[Dev,Prod]"}},
		{"trailing comma", "Owner,Env,", []string{"Owner", "Env", ""}},
		{"leading comma", ",Owner,Env", []string{"", "Owner", "Env"}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := splitTags(tc.input)
			if !reflect.DeepEqual(actual, tc.expected) {
				t.Errorf("splitTags(%q) = %v; want %v", tc.input, actual, tc.expected)
			}
		})
	}
}

func testParseTagsFatal(t *testing.T) {
	if os.Getenv("BE_TEST_FATAL") == "1" {
		parseTags("Invalid[Tag")
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=^TestParseTagsFatal$")
	cmd.Env = append(os.Environ(), "BE_TEST_FATAL=1")
	err := cmd.Run()
	if e, ok := err.(*exec.ExitError); ok && !e.Success() {
		return
	}
	t.Fatalf("process ran with err %v, want exit status 1", err)
}
