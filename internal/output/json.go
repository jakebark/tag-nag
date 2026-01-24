package output

import (
	"encoding/json"

	"github.com/jakebark/tag-nag/internal/shared"
)

// JSONFormatter implements JSON output format
type JSONFormatter struct{}

// JSONOutput represents the structured JSON output format
type JSONOutput struct {
	Violations []shared.Violation `json:"violations"`
	Summary    Summary           `json:"summary"`
}

// Summary provides aggregate information about violations
type Summary struct {
	Total         int `json:"total"`
	Skipped       int `json:"skipped"`
	FilesAffected int `json:"files_affected"`
}

// Format formats violations as JSON
func (f *JSONFormatter) Format(violations []shared.Violation) ([]byte, error) {
	// Calculate summary
	var skipped int
	files := make(map[string]bool)
	
	for _, v := range violations {
		if v.Skip {
			skipped++
		}
		files[v.FilePath] = true
	}
	
	output := JSONOutput{
		Violations: violations,
		Summary: Summary{
			Total:         len(violations),
			Skipped:       skipped,
			FilesAffected: len(files),
		},
	}
	
	return json.MarshalIndent(output, "", "  ")
}