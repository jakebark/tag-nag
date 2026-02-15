package output

import (
	"encoding/xml"
	"strings"
	"testing"

	"github.com/jakebark/tag-nag/internal/shared"
)

func TestJUnitXMLFormatter_Format(t *testing.T) {
	testCases := []struct {
		name         string
		violations   []shared.Violation
		wantXML      bool
		wantTests    int
		wantFailures int
	}{
		{
			name:         "empty violations",
			violations:   []shared.Violation{},
			wantXML:      true,
			wantTests:    0,
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
			wantXML:      true,
			wantTests:    1,
			wantFailures: 1,
		},
		{
			name: "multiple violations",
			violations: []shared.Violation{
				{ResourceType: "aws_s3_bucket", ResourceName: "test1", MissingTags: []string{"Owner"}},
				{ResourceType: "aws_instance", ResourceName: "test2", MissingTags: []string{"Env"}},
			},
			wantXML:      true,
			wantTests:    2,
			wantFailures: 2,
		},
		{
			name: "skipped violation",
			violations: []shared.Violation{
				{ResourceType: "aws_s3_bucket", ResourceName: "test", Skip: true},
			},
			wantXML:      true,
			wantTests:    1,
			wantFailures: 0,
		},
		{
			name: "mixed violations",
			violations: []shared.Violation{
				{ResourceType: "aws_s3_bucket", ResourceName: "test1", MissingTags: []string{"Owner"}},
				{ResourceType: "aws_instance", ResourceName: "test2", Skip: true},
			},
			wantXML:      true,
			wantTests:    2,
			wantFailures: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			formatter := &JUnitXMLFormatter{}
			output, err := formatter.Format(tc.violations)

			if err != nil {
				t.Errorf("Format() error = %v", err)
				return
			}

			if tc.wantXML {
				var testSuite TestSuite
				if err := xml.Unmarshal(output, &testSuite); err != nil {
					t.Errorf("Output is not valid XML: %v", err)
				}

				if testSuite.Tests != tc.wantTests {
					t.Errorf("Tests count = %d; want %d", testSuite.Tests, tc.wantTests)
				}

				if testSuite.Failures != tc.wantFailures {
					t.Errorf("Failures count = %d; want %d", testSuite.Failures, tc.wantFailures)
				}

				if !strings.Contains(string(output), "<?xml version") {
					t.Errorf("Output missing XML declaration")
				}
			}
		})
	}
}