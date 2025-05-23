package inputs

import (
	"reflect"
	"testing"

	"github.com/jakebark/tag-nag/internal/shared"
)

func TestParseTags(t *testing.T) {
	testCases := []struct {
		name          string
		input         string
		expected      shared.TagMap
		expectedError bool
	}{
		{
			name:  "key",
			input: "Owner",
			expected: shared.TagMap{
				"Owner": {},
			},
			expectedError: false,
		},
		{
			name:  "multiple keys",
			input: "Owner, Environment , Project",
			expected: shared.TagMap{
				"Owner":       {},
				"Environment": {},
				"Project":     {},
			},
			expectedError: false,
		},
		{
			name:  "mixed keys and values",
			input: "Owner[jake], Environment[Dev,Prod], CostCenter",
			expected: shared.TagMap{
				"Owner":       {"jake"},
				"Environment": {"Dev", "Prod"},
				"CostCenter":  {},
			},
			expectedError: false,
		},
		{
			name:          "empty",
			input:         "",
			expected:      shared.TagMap{},
			expectedError: false,
		},
		{
			name:          "whitespace",
			input:         "  ,   ",
			expected:      shared.TagMap{},
			expectedError: false,
		},
		{
			name:  "mixed keys and values, with whitespace",
			input: " Owner ,  Environment[Dev, Prod] ",
			expected: shared.TagMap{
				"Owner":       {},
				"Environment": {"Dev", "Prod"},
			},
			expectedError: false,
		},
		{
			name:  "leading comma",
			input: ",Owner",
			expected: shared.TagMap{
				"Owner": {},
			},
			expectedError: false,
		},
		{
			name:  "missing value",
			input: "Env[]",
			expected: shared.TagMap{
				"Env": {}, // No values extracted
			},
			expectedError: false,
		},
		{
			name:  "missing value, other values present",
			input: "Env[Dev,,Prod]",
			expected: shared.TagMap{
				"Env": {"Dev", "Prod"},
			},
			expectedError: false,
		},
		{
			name:  "whitespace preserved",
			input: "Owner[it belongs to me]",
			expected: shared.TagMap{
				"Owner": {"it belongs to me"},
			},
			expectedError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := parseTags(tc.input)
			if tc.expectedError {
				if err == nil {
					t.Errorf("parseTags(%q) expected an error, but got nil", tc.input)
				}
				return
			}
			if err != nil {
				t.Errorf("parseTags(%q) expected no error, but got: %v", tc.input, err)
			}
			if !reflect.DeepEqual(actual, tc.expected) {
				t.Errorf("parseTags(%q) = %v; want %v", tc.input, actual, tc.expected)
			}
		})
	}
}

func TestSplitTags(t *testing.T) {
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
