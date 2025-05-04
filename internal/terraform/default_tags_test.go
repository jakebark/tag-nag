package terraform

import (
	"testing"
)

// Test for normalizeProviderID
func testNormalizeProviderID(t *testing.T) {
	tests := []struct {
		name            string
		providerName    string
		alias           string
		caseInsensitive bool
		want            string
	}{
		{
			name:         "default",
			providerName: "aws",
			alias:        "",
			want:         "aws",
		},
		{
			name:         "alias",
			providerName: "aws",
			alias:        "west",
			want:         "aws.west",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := normalizeProviderID(tc.providerName, tc.alias, tc.caseInsensitive); got != tc.want {
				t.Errorf("normalizeProviderID() = %v, want %v", got, tc.want)
			}
		})
	}
}
