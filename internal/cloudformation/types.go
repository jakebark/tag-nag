package cloudformation

type Violation struct {
	ResourceName string
	ResourceType string
	Line         int
	MissingTags  []string
}

type TagMap map[string]string
