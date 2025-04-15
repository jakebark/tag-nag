package shared

const (
	TagNagIgnore    = "#tag-nag ignore"
	TagNagIgnoreAll = "#tag-nag ignore-all"
)

type TagMap map[string][]string

type Violation struct {
	FilePath     string
	ResourceType string
	ResourceName string
	Line         int
	MissingTags  []string
	Skip         bool
}
