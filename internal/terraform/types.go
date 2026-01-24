package terraform

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/jakebark/tag-nag/internal/shared"
)

type DefaultTags struct {
	LiteralTags map[string]shared.TagMap
}


type TerraformContext struct {
	EvalContext *hcl.EvalContext
}
