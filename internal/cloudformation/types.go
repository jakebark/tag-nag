package cloudformation

// Violation is a non-compliant tag
type Violation struct {
	ResourceName string
	ResourceType string
	Line         int
	MissingTags  []string
}

// TagMap maps tag keys to values
type TagMap map[string]string
