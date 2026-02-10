package output

import (
	"encoding/xml"
	"fmt"
	"strings"

	"github.com/jakebark/tag-nag/internal/shared"
)

// JUnitXMLFormatter implements JUnit XML output format
type JUnitXMLFormatter struct{}

// TestSuite represents a test suite in JUnit XML
type TestSuite struct {
	XMLName   xml.Name   `xml:"testsuite"`
	Name      string     `xml:"name,attr"`
	Tests     int        `xml:"tests,attr"`
	Failures  int        `xml:"failures,attr"`
	TestCases []TestCase `xml:"testcase"`
}

// TestCase represents a test case in JUnit XML
type TestCase struct {
	XMLName   xml.Name `xml:"testcase"`
	Name      string   `xml:"name,attr"`
	ClassName string   `xml:"classname,attr"`
	Failure   *Failure `xml:"failure,omitempty"`
}

// Failure represents a test failure in JUnit XML
type Failure struct {
	XMLName xml.Name `xml:"failure"`
	Message string   `xml:"message,attr"`
}

// Format formats violations as JUnit XML
func (f *JUnitXMLFormatter) Format(violations []shared.Violation) ([]byte, error) {
	var testCases []TestCase
	failures := 0

	for _, v := range violations {
		testCase := TestCase{
			Name:      fmt.Sprintf("%s.%s", v.ResourceType, v.ResourceName),
			ClassName: v.FilePath,
		}

		if !v.Skip {
			failures++
			testCase.Failure = &Failure{
				Message: fmt.Sprintf("Missing tags: %s", strings.Join(v.MissingTags, ", ")),
			}
		}

		testCases = append(testCases, testCase)
	}

	testSuite := TestSuite{
		Name:      "tag-nag",
		Tests:     len(violations),
		Failures:  failures,
		TestCases: testCases,
	}

	output, err := xml.MarshalIndent(testSuite, "", "  ")
	if err != nil {
		return nil, err
	}

	return []byte(xml.Header + string(output)), nil
}