package terraform

import "github.com/jakebark/tag-nag/internal/shared"

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
