package terraform

import "github.com/jakebark/tag-nag/internal/shared"

type DefaultTags struct {
	LiteralTags    map[string]shared.TagMap
	ReferencedTags TagReferences
}

// TagReferences maps a reference identifier to a tag map ("local.tags")
type TagReferences map[string]shared.TagMap
