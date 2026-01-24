package shared

type TagMap map[string][]string

type Violation struct {
	ResourceType string   `json:"resource_type"`
	ResourceName string   `json:"resource_name"`
	Line         int      `json:"line"`
	MissingTags  []string `json:"missing_tags"`
	Skip         bool     `json:"skip"`
	FilePath     string   `json:"file_path"`
}

type OutputFormat string

const (
	OutputFormatText OutputFormat = "text"
	OutputFormatJSON OutputFormat = "json"
)
