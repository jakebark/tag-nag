package cloudformation

// Violation is a non-compliant tag
type Violation struct {
	ResourceName string
	ResourceType string
	Line         int
	MissingTags  []string
}
