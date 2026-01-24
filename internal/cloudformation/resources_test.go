package cloudformation

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/jakebark/tag-nag/internal/shared"
)

func TestExtractTagMap(t *testing.T) {
	tests := []struct {
		name            string
		properties      map[string]any
		caseInsensitive bool
		expected        shared.TagMap
		expectedErr     bool
	}{
		{
			name: "no tag key",
			properties: map[string]any{
				"OtherKey": "Value"},
			expected: shared.TagMap{},
		},
		{
			name: "empty list",
			properties: map[string]any{
				"Tags": []any{},
			},
			expected: shared.TagMap{},
		},
		{
			name: "not a list",
			properties: map[string]any{
				"Tags": map[string]string{"Key": "Value"},
			},
			expectedErr: true,
		},
		{
			name: "literal tags",
			properties: map[string]any{
				"Tags": []any{
					map[string]any{"Key": "Owner", "Value": "Jake"},
					map[string]any{"Key": "Env", "Value": "Dev"},
				},
			},
			expected: shared.TagMap{
				"Owner": []string{"Jake"},
				"Env":   []string{"Dev"},
			},
		},
		// {
		// 	name: "referenced tags",
		// 	properties: map[string]any{
		// 		"Tags": []any{
		// 			map[string]any{"Key": "StackName", "Value": map[string]any{"Ref": "AWS::StackName"}},
		// 		},
		// 	},
		// 	expected: shared.TagMap{
		// 		"StackName": []string{"!Ref StackName"},
		// 	},
		// },
		// {
		// 	name: "mixed tags, literal and referenced",
		// 	properties: map[string]any{
		// 		"Tags": []any{
		// 			map[string]any{"Key": "Owner", "Value": "Jake"},
		// 			map[string]any{"Key": "StackName", "Value": map[string]any{"Ref": "AWS::StackName"}},
		// 		},
		// 	},
		// 	expected: shared.TagMap{
		// 		"Owner":     []string{"Jake"},
		// 		"StackName": []string{"!Ref StackName"},
		// 	},
		// },
		{
			name: "literal tags, case insensitive",
			properties: map[string]any{
				"Tags": []any{
					map[string]any{"Key": "Owner", "Value": "Jake"},
					map[string]any{"Key": "env", "Value": "Dev"},
				},
			},
			caseInsensitive: true,
			expected: shared.TagMap{
				"owner": []string{"Jake"},
				"env":   []string{"Dev"},
			},
		},
		{
			name: "missing key",
			properties: map[string]any{
				"Tags": []any{
					map[string]any{"Value": "Jake"},
				},
			},
			expected: shared.TagMap{},
		},
		{
			name: "missing value",
			properties: map[string]any{
				"Tags": []any{
					map[string]any{"Key": "OptionalTag"},
				},
			},
			expected: shared.TagMap{
				"OptionalTag": []string{""},
			},
		},
		{
			name: "non-string",
			properties: map[string]any{
				"Tags": []any{
					map[string]any{"Key": 123, "Value": "Jake"},
				},
			},
			expected: shared.TagMap{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := extractTagMap(tc.properties, tc.caseInsensitive)
			if (err != nil) != tc.expectedErr {
				t.Fatalf("extractTagMap() error = %v, expectedErr %v", err, tc.expectedErr)
			}
			if !tc.expectedErr {
				if diff := cmp.Diff(tc.expected, got); diff != "" {
					t.Errorf("extractTagMap() mismatch (-expected +got):\n%s", diff)
				}
			}
		})
	}
}
