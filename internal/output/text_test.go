package output

import (
	"strings"
	"testing"

	"github.com/jakebark/tag-nag/internal/shared"
)

func TestTextFormatter_Format(t *testing.T) {
	testCases := []struct {
		name           string
		violations     []shared.Violation
		wantContains   []string
		wantNotContains []string
	}{
		{
			name:       "empty violations",
			violations: []shared.Violation{},
			wantContains: []string{},
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
			wantContains: []string{
				"Violation(s) in main.tf",
				"10: aws_s3_bucket \"test\"",
				"Missing tags: Owner",
			},
		},
		{
			name: "multiple violations same file",
			violations: []shared.Violation{
				{ResourceType: "aws_s3_bucket", ResourceName: "test1", MissingTags: []string{"Owner"}, FilePath: "main.tf", Line: 5},
				{ResourceType: "aws_instance", ResourceName: "test2", MissingTags: []string{"Env"}, FilePath: "main.tf", Line: 15},
			},
			wantContains: []string{
				"Violation(s) in main.tf",
				"5: aws_s3_bucket \"test1\"",
				"15: aws_instance \"test2\"",
				"Missing tags: Owner",
				"Missing tags: Env",
			},
		},
		{
			name: "skipped violation",
			violations: []shared.Violation{
				{ResourceType: "aws_s3_bucket", ResourceName: "test", FilePath: "main.tf", Line: 10, Skip: true},
			},
			wantContains: []string{
				"Violation(s) in main.tf",
				"10: aws_s3_bucket \"test\" skipped",
			},
			wantNotContains: []string{
				"Missing tags:",
			},
		},
		{
			name: "mixed violations",
			violations: []shared.Violation{
				{ResourceType: "aws_s3_bucket", ResourceName: "test1", MissingTags: []string{"Owner"}, FilePath: "main.tf", Line: 5},
				{ResourceType: "aws_instance", ResourceName: "test2", FilePath: "main.tf", Line: 15, Skip: true},
			},
			wantContains: []string{
				"Violation(s) in main.tf",
				"5: aws_s3_bucket \"test1\"",
				"Missing tags: Owner",
				"15: aws_instance \"test2\" skipped",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			formatter := &TextFormatter{}
			output, err := formatter.Format(tc.violations)

			if err != nil {
				t.Errorf("Format() error = %v", err)
				return
			}

			outputStr := string(output)

			for _, want := range tc.wantContains {
				if !strings.Contains(outputStr, want) {
					t.Errorf("Output missing %q", want)
				}
			}

			for _, notWant := range tc.wantNotContains {
				if strings.Contains(outputStr, notWant) {
					t.Errorf("Output should not contain %q", notWant)
				}
			}
		})
	}
}