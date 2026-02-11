package output

import (
	"reflect"
	"testing"

	"github.com/jakebark/tag-nag/internal/shared"
)

func TestGetFormatter(t *testing.T) {
	testCases := []struct {
		name         string
		format       shared.OutputFormat
		expectedType string
	}{
		{
			name:         "json format",
			format:       shared.OutputFormatJSON,
			expectedType: "*output.JSONFormatter",
		},
		{
			name:         "junit-xml format",
			format:       shared.OutputFormatJUnitXML,
			expectedType: "*output.JUnitXMLFormatter",
		},
		{
			name:         "text format",
			format:       shared.OutputFormatText,
			expectedType: "*output.TextFormatter",
		},
		{
			name:         "unknown format defaults to text",
			format:       shared.OutputFormat("unknown"),
			expectedType: "*output.TextFormatter",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			formatter := GetFormatter(tc.format)
			actualType := reflect.TypeOf(formatter).String()
			if actualType != tc.expectedType {
				t.Errorf("GetFormatter(%q) = %s; want %s", tc.format, actualType, tc.expectedType)
			}
		})
	}
}

