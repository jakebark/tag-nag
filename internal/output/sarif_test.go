package output

import (
	"encoding/json"
	"testing"

	"github.com/jakebark/tag-nag/internal/shared"
)

func TestSARIFFormatter_Format(t *testing.T) {
	testCases := []struct {
		name         string
		violations   []shared.Violation
		wantResults  int
		wantFailures int
	}{
		{
			name:         "empty violations",
			violations:   []shared.Violation{},
			wantResults:  0,
			wantFailures: 0,
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
			wantResults:  1,
			wantFailures: 1,
		},
		{
			name: "multiple violations",
			violations: []shared.Violation{
				{ResourceType: "aws_s3_bucket", ResourceName: "test1", MissingTags: []string{"Owner"}, FilePath: "main.tf", Line: 1},
				{ResourceType: "aws_instance", ResourceName: "test2", MissingTags: []string{"Env"}, FilePath: "main.tf", Line: 20},
			},
			wantResults:  2,
			wantFailures: 2,
		},
		{
			name: "skipped violation",
			violations: []shared.Violation{
				{ResourceType: "aws_s3_bucket", ResourceName: "test", Skip: true, FilePath: "main.tf", Line: 1},
			},
			wantResults:  1,
			wantFailures: 0,
		},
		{
			name: "mixed violations",
			violations: []shared.Violation{
				{ResourceType: "aws_s3_bucket", ResourceName: "test1", MissingTags: []string{"Owner"}, FilePath: "main.tf", Line: 1},
				{ResourceType: "aws_instance", ResourceName: "test2", Skip: true, FilePath: "main.tf", Line: 10},
			},
			wantResults:  2,
			wantFailures: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			formatter := &SARIFFormatter{}
			output, err := formatter.Format(tc.violations)

			if err != nil {
				t.Errorf("Format() error = %v", err)
				return
			}

			var parsed sarifOutput
			if err := json.Unmarshal(output, &parsed); err != nil {
				t.Errorf("Output is not valid JSON: %v", err)
				return
			}

			if parsed.Version != "2.1.0" {
				t.Errorf("Version = %q; want %q", parsed.Version, "2.1.0")
			}

			if len(parsed.Runs) != 1 {
				t.Fatalf("Runs count = %d; want 1", len(parsed.Runs))
			}

			results := parsed.Runs[0].Results
			if len(results) != tc.wantResults {
				t.Errorf("Results count = %d; want %d", len(results), tc.wantResults)
			}

			failures := 0
			for _, r := range results {
				if r.Kind == "fail" {
					failures++
				}
			}
			if failures != tc.wantFailures {
				t.Errorf("Failure count = %d; want %d", failures, tc.wantFailures)
			}
		})
	}
}
