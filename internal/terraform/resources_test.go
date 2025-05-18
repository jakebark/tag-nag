package terraform

import (
	"sort"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/jakebark/tag-nag/internal/shared"
	"github.com/zclconf/go-cty/cty"
)

func TestCheckResourcesForTags_Taggability(t *testing.T) {
	parser := hclparse.NewParser()

	tfCode := `
		resource "aws_s3_bucket" "taggable_bucket" {
		  tags = {
			Owner = "test-user" // Missing Environment
		  }
		}

		resource "aws_kms_alias" "non_taggable_alias" {
		  alias_name       = "alias/my-key-alias"
		  target_key_id = "some-key-id"
		}

		resource "aws_instance" "another_taggable" {
		  ami           = "ami-12345"
		  instance_type = "t2.micro"
		  tags = {
			Owner = "test-user"
			Environment = "dev"
		  }
		}

		resource "aws_route53_zone" "unknown_in_schema_should_be_checked" {
			name = "example.com"
			# Missing all tags
		}
	`

	file, diags := parser.ParseHCL([]byte(tfCode), "test.tf")
	if diags.HasErrors() {
		t.Fatalf("Failed to parse test HCL: %v", diags)
	}

	body, ok := file.Body.(*hclsyntax.Body)
	if !ok {
		t.Fatalf("Could not get HCL syntax body")
	}

	requiredTags := shared.TagMap{
		"Owner":       {},
		"Environment": {},
	}
	lines := strings.Split(tfCode, "\n")

	t.Run("With Taggability Filter", func(t *testing.T) {
		taggableMap := map[string]bool{
			"aws_s3_bucket":    true,
			"aws_kms_alias":    false,
			"aws_instance":     true,
			"aws_iam_user":     false, // Example of another non-taggable not in the HCL
			"aws_route53_zone": true,  // Explicitly mark as taggable for test
		}

		mockCtx := &TerraformContext{EvalContext: &hcl.EvalContext{Variables: make(map[string]cty.Value)}}
		mockDefaults := &DefaultTags{LiteralTags: make(map[string]shared.TagMap)}

		violations := checkResourcesForTags(body, requiredTags, mockDefaults, mockCtx, false, lines, false, taggableMap)

		expectedViolations := []Violation{
			{resourceType: "aws_s3_bucket", resourceName: "taggable_bucket", line: 2, missingTags: []string{"Environment"}},
			{resourceType: "aws_route53_zone", resourceName: "unknown_in_schema_should_be_checked", line: 19, missingTags: []string{"Environment", "Owner"}}, // Order might vary
		}

		sortViolations(violations)
		sortViolations(expectedViolations) // Sort missing tags within each violation for stable comparison

		if diff := cmp.Diff(expectedViolations, violations, cmpopts.IgnoreUnexported(Violation{})); diff != "" {
			t.Errorf("checkResourcesForTags with filter mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("Without Taggability Filter (nil map)", func(t *testing.T) {
		taggableMap := map[string]bool(nil)

		mockCtx := &TerraformContext{EvalContext: &hcl.EvalContext{Variables: make(map[string]cty.Value)}}
		mockDefaults := &DefaultTags{LiteralTags: make(map[string]shared.TagMap)}

		violations := checkResourcesForTags(body, requiredTags, mockDefaults, mockCtx, false, lines, false, taggableMap)

		expectedViolations := []Violation{
			{resourceType: "aws_s3_bucket", resourceName: "taggable_bucket", line: 2, missingTags: []string{"Environment"}},
			{resourceType: "aws_kms_alias", resourceName: "non_taggable_alias", line: 8, missingTags: []string{"Environment", "Owner"}},
			{resourceType: "aws_route53_zone", resourceName: "unknown_in_schema_should_be_checked", line: 19, missingTags: []string{"Environment", "Owner"}},
		}
		sortViolations(violations)
		sortViolations(expectedViolations)

		if diff := cmp.Diff(expectedViolations, violations, cmpopts.IgnoreUnexported(Violation{})); diff != "" {
			t.Errorf("checkResourcesForTags without filter mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("With Taggability Filter - resource type not in map (should assume taggable)", func(t *testing.T) {
		// aws_route53_zone is NOT in this specific taggableMap
		taggableMap := map[string]bool{
			"aws_s3_bucket": true,
			"aws_kms_alias": false,
			"aws_instance":  true,
		}

		mockCtx := &TerraformContext{EvalContext: &hcl.EvalContext{Variables: make(map[string]cty.Value)}}
		mockDefaults := &DefaultTags{LiteralTags: make(map[string]shared.TagMap)}

		violations := checkResourcesForTags(body, requiredTags, mockDefaults, mockCtx, false, lines, false, taggableMap)

		expectedViolations := []Violation{
			{resourceType: "aws_s3_bucket", resourceName: "taggable_bucket", line: 2, missingTags: []string{"Environment"}},
			// aws_route53_zone is assumed taggable as it's not in the map with a 'false' entry
			{resourceType: "aws_route53_zone", resourceName: "unknown_in_schema_should_be_checked", line: 19, missingTags: []string{"Environment", "Owner"}},
		}

		sortViolations(violations)
		sortViolations(expectedViolations)

		if diff := cmp.Diff(expectedViolations, violations, cmpopts.IgnoreUnexported(Violation{})); diff != "" {
			t.Errorf("checkResourcesForTags with incomplete filter mismatch (-want +got):\n%s", diff)
		}
	})
}

// Helper to sort violations for consistent comparison
func sortViolations(violations []Violation) {
	for i := range violations {
		sort.Strings(violations[i].missingTags)
	}
	sort.Slice(violations, func(i, j int) bool {
		if violations[i].resourceType != violations[j].resourceType {
			return violations[i].resourceType < violations[j].resourceType
		}
		return violations[i].resourceName < violations[j].resourceName
	})
}
