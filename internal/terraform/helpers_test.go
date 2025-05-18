package terraform

import (
	"reflect"
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/jakebark/tag-nag/internal/shared"
	"github.com/zclconf/go-cty/cty"
)

func TestTraversalToString(t *testing.T) {
	testCases := []struct {
		name            string
		hclInput        string
		caseInsensitive bool
		expected        string
	}{
		{
			name:            "literal",
			hclInput:        "Owner",
			caseInsensitive: false,
			expected:        "Owner",
		},
		{
			name:            "literal, case insensitive",
			hclInput:        "Owner",
			caseInsensitive: true,
			expected:        "owner",
		},
		{
			name:            "traversal",
			hclInput:        "local.network.subnets[0].id",
			caseInsensitive: false,
			expected:        "local.network.subnets",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Parse the HCL expression string
			expr, diags := hclsyntax.ParseExpression([]byte(tc.hclInput), tc.name+".tf", hcl.Pos{Line: 1, Column: 1})
			if diags.HasErrors() {
				t.Fatalf("Failed to parse expression %q: %v", tc.hclInput, diags)
			}

			actual := traversalToString(expr, tc.caseInsensitive)
			if actual != tc.expected {
				t.Errorf("traversalToString(%q, %v) = %q; want %q", tc.hclInput, tc.caseInsensitive, actual, tc.expected)
			}
		})
	}
}

func TestMergeTags(t *testing.T) {
	testCases := []struct {
		name     string
		inputs   []shared.TagMap
		expected shared.TagMap
	}{
		{
			name:     "empty",
			inputs:   []shared.TagMap{},
			expected: shared.TagMap{},
		},
		{
			name: "key",
			inputs: []shared.TagMap{
				{"Environment": {}},
			},
			expected: shared.TagMap{"Environment": {}},
		},
		{
			name:     "key and value",
			inputs:   []shared.TagMap{{"Environment": {"Dev"}}},
			expected: shared.TagMap{"Environment": {"Dev"}},
		},
		{
			name: "multiple keys and values",
			inputs: []shared.TagMap{
				{"Environment": {"Dev"}},
				{"Owner": {"Prod"}},
			},
			expected: shared.TagMap{"Environment": {"Dev"}, "Owner": {"Prod"}},
		},
		{
			name: "overlapping values, last wins",
			inputs: []shared.TagMap{
				{"Environment": {"Dev"}, "Owner": {"jakebark"}},
				{"Owner": {"Jake"}, "CostCenter": {"C-01"}},
			},
			expected: shared.TagMap{"Environment": {"Dev"}, "Owner": {"Jake"}, "CostCenter": {"C-01"}},
		},
		{
			name: "overlapping empty value, last wins",
			inputs: []shared.TagMap{
				{"Environment": {"Dev"}},
				{"Environment": {}},
			},
			expected: shared.TagMap{"Environment": {}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := mergeTags(tc.inputs...)
			if !reflect.DeepEqual(actual, tc.expected) {
				t.Errorf("mergeTags() = %v; want %v", actual, tc.expected)
			}
		})
	}
}

func TestConvertCtyValueToString(t *testing.T) {
	tests := []struct {
		name    string
		input   cty.Value
		want    string
		wantErr bool
	}{
		{
			name:  "string",
			input: cty.StringVal("hello world"),
			want:  "hello world",
		},
		{
			name:  "empty string",
			input: cty.StringVal(""),
			want:  "",
		},
		{
			name:  "number",
			input: cty.NumberIntVal(123),
			want:  "123",
		},
		{
			name:  "float",
			input: cty.NumberFloatVal(123.45),
			want:  "123.45",
		},
		{
			name:  "bool, true",
			input: cty.True,
			want:  "true",
		},
		{
			name:  "bool, false",
			input: cty.False,
			want:  "false",
		},
		{
			name:  "null",
			input: cty.NullVal(cty.String),
			want:  "",
		},
		{
			name:    "unknown value",
			input:   cty.UnknownVal(cty.String),
			wantErr: true,
		},
		{
			name:  "list",
			input: cty.ListVal([]cty.Value{cty.StringVal("a"), cty.StringVal("b")}),
			want:  `["a","b"]`,
		},
		{
			name:  "map",
			input: cty.MapVal(map[string]cty.Value{"key": cty.StringVal("value")}),
			want:  `{"key":"value"}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := convertCtyValueToString(tc.input)
			if (err != nil) != tc.wantErr {
				t.Fatalf("convertCtyValueToString() error = %v, wantErr %v", err, tc.wantErr)
			}
			if !tc.wantErr && got != tc.want {
				t.Errorf("convertCtyValueToString() = %v, want %v", got, tc.want)
			}
		})
	}
}
