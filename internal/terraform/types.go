package terraform

// DefaultTags hold the default_tag values
type DefaultTags struct {
	LiteralTags    map[string]TagMap
	ReferencedTags TagReferences
}

// Violation is a non-compliant tag
type Violation struct {
	resourceType string
	resourceName string
	line         int
	missingTags  []string
}

// TagMap maps tag keys to values
type TagMap map[string]string

// TagReferences maps a reference identifier to a tag map ("local.tags")
type TagReferences map[string]TagMap
