package shared

import (
	"reflect"
	"sort"
	"testing"
)

func TestFilterMissingTags(t *testing.T) {
	testCases := []struct {
		name            string
		requiredTags    TagMap
		effectiveTags   TagMap
		caseInsensitive bool
		expectedMissing []string
	}{
		{
			name:            "tags present",
			requiredTags:    TagMap{"Owner": {}, "Env": {"Prod", "Dev"}},
			effectiveTags:   TagMap{"Owner": {"a"}, "Env": {"Prod"}},
			caseInsensitive: false,
			expectedMissing: nil,
		},
		{
			name:            "missing key",
			requiredTags:    TagMap{"Owner": {}, "Env": {"Prod"}},
			effectiveTags:   TagMap{"Env": {"Prod"}},
			caseInsensitive: false,
			expectedMissing: []string{"Owner"},
		},
		{
			name:            "wrong value",
			requiredTags:    TagMap{"Owner": {}, "Env": {"Prod"}},
			effectiveTags:   TagMap{"Owner": {"a"}, "Env": {"Dev"}},
			caseInsensitive: false,
			expectedMissing: []string{"Env[Prod]"},
		},
		{
			name:            "missing key and value",
			requiredTags:    TagMap{"Env": {"Prod"}},
			effectiveTags:   TagMap{"Owner": {"a"}},
			caseInsensitive: false,
			expectedMissing: []string{"Env[Prod]"},
		},
		{
			name:            "tags present, case insensitive",
			requiredTags:    TagMap{"Owner": {}, "Env": {"Prod", "Dev"}},
			effectiveTags:   TagMap{"owner": {"a"}, "env": {"prod"}},
			caseInsensitive: true,
			expectedMissing: nil,
		},
		{
			name:            "missing key, case insensitive",
			requiredTags:    TagMap{"Owner": {}, "Env": {"Prod"}},
			effectiveTags:   TagMap{"env": {"Prod"}},
			caseInsensitive: true,
			expectedMissing: []string{"Owner"},
		},
		{
			name:            "wrong value, case insensitive",
			requiredTags:    TagMap{"Owner": {}, "Env": {"Prod"}},
			effectiveTags:   TagMap{"owner": {"a"}, "env": {"Dev"}},
			caseInsensitive: true,
			expectedMissing: []string{"Env[Prod]"},
		},
		{
			name:            "missing key and value, case insensitive",
			requiredTags:    TagMap{"Env": {"Prod"}},
			effectiveTags:   TagMap{"owner": {"a"}},
			caseInsensitive: true,
			expectedMissing: []string{"Env[Prod]"},
		},
		{
			name:            "no required tags",
			requiredTags:    TagMap{},
			effectiveTags:   TagMap{"Owner": {"a"}, "Env": {"Dev"}},
			caseInsensitive: false,
			expectedMissing: nil,
		},
		{
			name:            "no tags",
			requiredTags:    TagMap{"Owner": {}, "Env": {"Prod"}},
			effectiveTags:   TagMap{},
			caseInsensitive: false,
			expectedMissing: []string{"Owner", "Env[Prod]"},
		},
		{
			name:            "multiple values required, one present",
			requiredTags:    TagMap{"Region": {"us-east-1", "us-west-2"}},
			effectiveTags:   TagMap{"Region": {"us-west-2"}},
			caseInsensitive: false,
			expectedMissing: nil,
		},
		{
			name:            "multiple values required, none present",
			requiredTags:    TagMap{"Region": {"us-east-1", "us-west-2"}},
			effectiveTags:   TagMap{"Region": {"eu-central-1"}},
			caseInsensitive: false,
			expectedMissing: []string{"Region[us-east-1,us-west-2]"},
		},
		{
			name:            "multiple values required, one present, case insensitive",
			requiredTags:    TagMap{"Env": {"Prod", "Dev"}},
			effectiveTags:   TagMap{"Env": {"prod"}},
			caseInsensitive: true,
			expectedMissing: nil,
		},
		{
			name:            "multiple values required, none present, case insensitive",
			requiredTags:    TagMap{"Region": {"us-east-1", "us-west-2"}},
			effectiveTags:   TagMap{"region": {"eu-central-1"}},
			caseInsensitive: true,
			expectedMissing: []string{"Region[us-east-1,us-west-2]"},
		},
		{
			name:            "multiple values required, multiple values present",
			requiredTags:    TagMap{"Env": {"Prod", "Dev"}},
			effectiveTags:   TagMap{"Env": {"Prod", "Stage"}},
			caseInsensitive: false,
			expectedMissing: nil,
		},
		{
			name:            "multiple values present, none match",
			requiredTags:    TagMap{"Env": {"Prod"}},
			effectiveTags:   TagMap{"Env": {"Dev", "Stage"}},
			caseInsensitive: false,
			expectedMissing: []string{"Env[Prod]"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := FilterMissingTags(tc.requiredTags, tc.effectiveTags, tc.caseInsensitive)

			if actual != nil && tc.expectedMissing != nil {
				sort.Strings(actual)
				sort.Strings(tc.expectedMissing)
			}

			if !reflect.DeepEqual(actual, tc.expectedMissing) {
				t.Errorf("FilterMissingTags() = %#v; want %#v", actual, tc.expectedMissing)
			}
		})
	}
}

func TestNormalizeCase(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		caseInsensitive bool
		expected        string
	}{
		{
			name:            "preserve case when false",
			input:           "Owner",
			caseInsensitive: false,
			expected:        "Owner",
		},
		{
			name:            "lowercase when true",
			input:           "Owner",
			caseInsensitive: true,
			expected:        "owner",
		},
		{
			name:            "already lowercase",
			input:           "owner",
			caseInsensitive: true,
			expected:        "owner",
		},
		{
			name:            "mixed case",
			input:           "EnViRoNmEnT",
			caseInsensitive: true,
			expected:        "environment",
		},
		{
			name:            "all uppercase",
			input:           "ENVIRONMENT",
			caseInsensitive: true,
			expected:        "environment",
		},
		{
			name:            "empty string case sensitive",
			input:           "",
			caseInsensitive: false,
			expected:        "",
		},
		{
			name:            "empty string case insensitive",
			input:           "",
			caseInsensitive: true,
			expected:        "",
		},
		{
			name:            "special characters preserved",
			input:           "Tag-Name_123",
			caseInsensitive: true,
			expected:        "tag-name_123",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := NormalizeCase(tc.input, tc.caseInsensitive)
			if actual != tc.expected {
				t.Errorf("NormalizeCase(%q, %v) = %q; want %q", tc.input, tc.caseInsensitive, actual, tc.expected)
			}
		})
	}
}

func TestCompareCase(t *testing.T) {
	tests := []struct {
		name            string
		first           string
		second          string
		caseInsensitive bool
		expected        bool
	}{
		{
			name:            "exact match case sensitive",
			first:           "Owner",
			second:          "Owner",
			caseInsensitive: false,
			expected:        true,
		},
		{
			name:            "case mismatch case sensitive",
			first:           "Owner",
			second:          "owner",
			caseInsensitive: false,
			expected:        false,
		},
		{
			name:            "case mismatch case insensitive",
			first:           "Owner",
			second:          "owner",
			caseInsensitive: true,
			expected:        true,
		},
		{
			name:            "different strings case sensitive",
			first:           "Owner",
			second:          "Environment",
			caseInsensitive: false,
			expected:        false,
		},
		{
			name:            "different strings case insensitive",
			first:           "Owner",
			second:          "Environment",
			caseInsensitive: true,
			expected:        false,
		},
		{
			name:            "mixed case match case insensitive",
			first:           "EnViRoNmEnT",
			second:          "environment",
			caseInsensitive: true,
			expected:        true,
		},
		{
			name:            "both uppercase case insensitive",
			first:           "ENVIRONMENT",
			second:          "ENVIRONMENT",
			caseInsensitive: true,
			expected:        true,
		},
		{
			name:            "empty strings",
			first:           "",
			second:          "",
			caseInsensitive: false,
			expected:        true,
		},
		{
			name:            "empty vs non-empty",
			first:           "",
			second:          "Owner",
			caseInsensitive: true,
			expected:        false,
		},
		{
			name:            "special characters case insensitive",
			first:           "Tag-Name_123",
			second:          "tag-name_123",
			caseInsensitive: true,
			expected:        true,
		},
		{
			name:            "unicode case insensitive",
			first:           "Café",
			second:          "café",
			caseInsensitive: true,
			expected:        true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := CompareCase(tc.first, tc.second, tc.caseInsensitive)
			if actual != tc.expected {
				t.Errorf("CompareCase(%q, %q, %v) = %v; want %v", tc.first, tc.second, tc.caseInsensitive, actual, tc.expected)
			}
		})
	}
}
