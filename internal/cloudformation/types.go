package cloudformation

type Violation struct {
	resourceName string
	resourceType string
	line         int
	missingTags  []string
	skip         bool
}
