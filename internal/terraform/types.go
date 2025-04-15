package terraform

import "github.com/jakebark/tag-nag/internal/shared"

type DefaultTags struct {
	LiteralTags    map[string]shared.TagMap
	ReferencedTags TagReferences
}

type Violation struct {
	resourceType string
	resourceName string
	line         int
	missingTags  []string
	skip         bool
}

// TagReferences maps a reference identifier to a tag map ("local.tags")
type TagReferences map[string]shared.TagMap
