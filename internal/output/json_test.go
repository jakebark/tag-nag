package output

import (
	"encoding/json"
	"testing"

	"github.com/jakebark/tag-nag/internal/shared"
)

func TestJSONFormatter_Format(t *testing.T) {
	testCases := []struct {
		name       string
		violations []shared.Violation
		wantJSON   bool
		wantFields []string
	}{
		{
			name:       "empty violations",
			violations: []shared.Violation{},
			wantJSON:   true,
			wantFields: []string{"violations", "summary"},
		},
		{
			name: "single violation",
			violations: []shared.Violation{
				{
					ResourceType: "aws_s3_bucket",
					ResourceName: "test",
					MissingTags:  []string{"Owner"},
					FilePath:     "main.tf",
					Line:         10,
				},
			},
			wantJSON:   true,
			wantFields: []string{"violations", "summary"},
		},
		{
			name: "multiple violations",
			violations: []shared.Violation{
				{ResourceType: "aws_s3_bucket", ResourceName: "test1", MissingTags: []string{"Owner"}},
				{ResourceType: "aws_instance", ResourceName: "test2", MissingTags: []string{"Env"}},
			},
			wantJSON:   true,
			wantFields: []string{"violations", "summary"},
		},
		{
			name: "skipped violation",
			violations: []shared.Violation{
				{ResourceType: "aws_s3_bucket", ResourceName: "test", Skip: true},
			},
			wantJSON:   true,
			wantFields: []string{"violations", "summary"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			formatter := &JSONFormatter{}
			output, err := formatter.Format(tc.violations)

			if err != nil {
				t.Errorf("Format() error = %v", err)
				return
			}

			if tc.wantJSON {
				var parsed map[string]interface{}
				if err := json.Unmarshal(output, &parsed); err != nil {
					t.Errorf("Output is not valid JSON: %v", err)
				}

				for _, field := range tc.wantFields {
					if _, exists := parsed[field]; !exists {
						t.Errorf("Missing field %q in JSON output", field)
					}
				}
			}
		})
	}
}