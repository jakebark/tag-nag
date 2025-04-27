package terraform

import (
	"reflect"
	"testing"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/jakebark/tag-nag/internal/shared"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/json"
)

func testTraversalToString(t *testing.T) {
	testCases := []struct {
		name            string
		hclInput        string
		caseInsensitive bool
		expected        string
	}{
		{
			name:            "literal",
			hclInput:        `Owner`,
			caseInsensitive: false,
			expected:        "Owner",
		},
		{
			name:            "literal, case insensitive",
			hclInput:        `Owner`,
			caseInsensitive: true,
			expected:        "owner",
		},
		{
			name:            "traversal",
			hclInput:        `local.network.subnets[0].id`,
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
