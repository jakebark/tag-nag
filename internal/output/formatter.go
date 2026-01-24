package output

import "github.com/jakebark/tag-nag/internal/shared"

// Formatter defines the interface for all output formatters
type Formatter interface {
	Format(violations []shared.Violation) ([]byte, error)
}

// GetFormatter returns the appropriate formatter for the given format
func GetFormatter(format shared.OutputFormat) Formatter {
	switch format {
	case shared.OutputFormatJSON:
		return &JSONFormatter{}
	case shared.OutputFormatText:
		fallthrough
	default:
		return &TextFormatter{}
	}
}