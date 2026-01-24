package output

import (
	"fmt"
	"strings"

	"github.com/jakebark/tag-nag/internal/shared"
)

// TextFormatter implements the current text output format
type TextFormatter struct{}

// Format formats violations as human-readable text
func (f *TextFormatter) Format(violations []shared.Violation) ([]byte, error) {
	var output strings.Builder
	
	// Group violations by file
	fileGroups := make(map[string][]shared.Violation)
	for _, v := range violations {
		fileGroups[v.FilePath] = append(fileGroups[v.FilePath], v)
	}
	
	// Format each file group
	for filePath, fileViolations := range fileGroups {
		output.WriteString(fmt.Sprintf("\nViolation(s) in %s\n", filePath))
		
		for _, v := range fileViolations {
			if v.Skip {
				output.WriteString(fmt.Sprintf("  %d: %s \"%s\" skipped\n", 
					v.Line, v.ResourceType, v.ResourceName))
			} else {
				output.WriteString(fmt.Sprintf("  %d: %s \"%s\" 🏷️  Missing tags: %s\n", 
					v.Line, v.ResourceType, v.ResourceName, strings.Join(v.MissingTags, ", ")))
			}
		}
	}
	
	return []byte(output.String()), nil
}