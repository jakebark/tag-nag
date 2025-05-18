package terraform

import (
	"testing"
)

// Test for normalizeProviderID
func TestNormalizeProviderID(t *testing.T) {
	tests := []struct {
		name            string
		providerName    string
		alias           string
		caseInsensitive bool
		expected        string
	}{
		{
			name:         "default",
			providerName: "aws",
			alias:        "",
			expected:     "aws",
		},
		{
			name:         "alias",
			providerName: "aws",
			alias:        "west",
			expected:     "aws.west",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := normalizeProviderID(tc.providerName, tc.alias, tc.caseInsensitive); got != tc.expected {
				t.Errorf("normalizeProviderID() = %v, expected %v", got, tc.expected)
			}
		})
	}
}
