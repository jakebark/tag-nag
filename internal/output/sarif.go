package output

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jakebark/tag-nag/internal/shared"
)

type SARIFFormatter struct{}

type sarifOutput struct {
	Schema  string     `json:"$schema"`
	Version string     `json:"version"`
	Runs    []sarifRun `json:"runs"`
}

type sarifRun struct {
	Tool struct {
		Driver struct {
			Name           string `json:"name"`
			InformationURI string `json:"informationUri"`
		} `json:"driver"`
	} `json:"tool"`
	Results []sarifResult `json:"results"`
}

type sarifResult struct {
	RuleID  string `json:"ruleId"`
	Kind    string `json:"kind"`
	Level   string `json:"level,omitempty"`
	Message struct {
		Text string `json:"text"`
	} `json:"message"`
	Locations []sarifLocation `json:"locations"`
}

type sarifLocation struct {
	PhysicalLocation struct {
		ArtifactLocation struct {
			URI string `json:"uri"`
		} `json:"artifactLocation"`
		Region struct {
			StartLine int `json:"startLine"`
		} `json:"region"`
	} `json:"physicalLocation"`
}

// Format formats violations as SARIF v2.1.0
func (f *SARIFFormatter) Format(violations []shared.Violation) ([]byte, error) {
	var results []sarifResult

	for _, v := range violations {
		r := sarifResult{RuleID: "missing-tags"}
		r.Message.Text = fmt.Sprintf("%s %q is missing tags: %s",
			v.ResourceType, v.ResourceName, strings.Join(v.MissingTags, ", "))

		if v.Skip {
			r.Kind = "notApplicable"
		} else {
			r.Kind = "fail"
			r.Level = "error"
		}

		loc := sarifLocation{}
		loc.PhysicalLocation.ArtifactLocation.URI = v.FilePath
		loc.PhysicalLocation.Region.StartLine = v.Line
		r.Locations = []sarifLocation{loc}

		results = append(results, r)
	}

	run := sarifRun{Results: results}
	run.Tool.Driver.Name = "tag-nag"
	run.Tool.Driver.InformationURI = "https://github.com/jakebark/tag-nag"

	output := sarifOutput{
		Schema:  "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/main/sarif-2.1/schema/sarif-schema-2.1.0.json",
		Version: "2.1.0",
		Runs:    []sarifRun{run},
	}

	return json.MarshalIndent(output, "", "  ")
}
