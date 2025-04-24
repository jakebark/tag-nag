package terraform

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/jakebark/tag-nag/internal/shared"
)

type DefaultTags struct {
	LiteralTags map[string]shared.TagMap
}

type Violation struct {
	resourceType string
	resourceName string
	line         int
	missingTags  []string
	skip         bool
}

type TerraformContext struct {
	EvalContext *hcl.EvalContext
}
