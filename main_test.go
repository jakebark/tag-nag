package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
)

const binaryName = "tag-nag"

type testCases struct {
	name             string
	filePathOrDir    string
	cliArgs          []string
	expectedExitCode int
	expectedError    bool
	expectedOutput   []string
}

func TestMain(m *testing.M) {
	cmd := exec.Command("go", "build", "-o", binaryName)
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to build %s: %v\n", binaryName, err)
		os.Exit(1)
	}

	exitVal := m.Run()
	os.Remove(binaryName)
	os.Exit(exitVal)
}

func runTagNag(t *testing.T, args ...string) (string, error, int) {
	t.Helper()
	fullArgs := append([]string{"./" + binaryName}, args...)
	cmd := exec.Command(fullArgs[0], fullArgs[1:]...)
	var outbuf, errbuf bytes.Buffer
	cmd.Stdout = &outbuf
	cmd.Stderr = &errbuf

	err := cmd.Run()
	stdout := outbuf.String()
	stderr := errbuf.String()

	fullOutput := stdout
	if stderr != "" {
		fullOutput += "\n" + stderr
	}

	exitCode := 0
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		} else {
			t.Fatalf("Command execution failed with non-exit error: %v, output: %s", err, fullOutput)
		}
	}
	return fullOutput, err, exitCode
}

func TestTerraformCLI(t *testing.T) {
	testCases := []testCases{
		{
			name:             "tags",
			filePathOrDir:    "testdata/terraform/single_resource.tf",
			cliArgs:          []string{"--tags", "Owner,Environment"},
			expectedExitCode: 0,
			expectedError:    false,
			expectedOutput:   []string{"No tag violations found"},
		},
		{
			name:             "missing tags",
			filePathOrDir:    "testdata/terraform/single_resource.tf",
			cliArgs:          []string{"--tags", "Owner,Environment,Project"},
			expectedExitCode: 1,
			expectedError:    true,
			expectedOutput:   []string{`aws_s3_bucket "this"`, "Missing tags: Project"},
		},
		{
			name:             "no tags",
			filePathOrDir:    "testdata/terraform/no_tags.tf",
			cliArgs:          []string{"--tags", "Owner,Environment"},
			expectedExitCode: 1,
			expectedError:    true,
			expectedOutput:   []string{`aws_s3_bucket "this"`, "Missing tags: Owner, Environment"},
		},
		{
			name:             "case insensitive",
			filePathOrDir:    "testdata/terraform/single_resource.tf",
			cliArgs:          []string{"--tags", "owner,environment", "-c"},
			expectedExitCode: 0,
			expectedError:    false,
			expectedOutput:   []string{"No tag violations found"},
		},
		{
			name:             "lower case",
			filePathOrDir:    "testdata/terraform/single_resource.tf",
			cliArgs:          []string{"--tags", "owner,environment"},
			expectedExitCode: 1,
			expectedError:    true,
			expectedOutput:   []string{`aws_s3_bucket "this"`, "Missing tags: owner, environment"},
		},
		{
			name:             "tag values",
			filePathOrDir:    "testdata/terraform/single_resource.tf",
			cliArgs:          []string{"--tags", "Owner,Environment[dev,prod]"},
			expectedExitCode: 0,
			expectedError:    false,
			expectedOutput:   []string{"No tag violations found"},
		},
		{
			name:             "missing tag value",
			filePathOrDir:    "testdata/terraform/single_resource.tf",
			cliArgs:          []string{"--tags", "Owner,Environment[test]"},
			expectedExitCode: 1,
			expectedError:    true,
			expectedOutput:   []string{`aws_s3_bucket "this"`, "Missing tags: Environment[test]"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			argsForRun := append([]string{tc.filePathOrDir}, tc.cliArgs...)
			output, err, exitCode := runTagNag(t, argsForRun...)

			if tc.expectedError && err == nil {
				t.Errorf("Expected an error from command execution, but got none. Output:\n%s", output)
			}
			if !tc.expectedError && err != nil {
				t.Errorf("Expected no error from command execution, but got: %v. Output:\n%s", err, output)
			}

			if exitCode != tc.expectedExitCode {
				t.Errorf("Expected exit code %d, got %d. Output:\n%s", tc.expectedExitCode, exitCode, output)
			}

			for _, expectedStr := range tc.expectedOutput {
				if !strings.Contains(output, expectedStr) {
					t.Errorf("Output missing expected string '%s'. Output:\n%s", expectedStr, output)
				}
			}
		})
	}
}

func TestCloudFormationCLI(t *testing.T) {
	testCases := []testCases{
		{
			name:             "yml",
			filePathOrDir:    "testdata/cloudformation/single_resource.yml",
			cliArgs:          []string{"--tags", "Owner,Environment"},
			expectedExitCode: 0,
			expectedError:    false,
			expectedOutput:   []string{"No tag violations found"},
		},
		{
			name:             "yaml",
			filePathOrDir:    "testdata/cloudformation/single_resource.yaml",
			cliArgs:          []string{"--tags", "Owner,Environment"},
			expectedExitCode: 0,
			expectedError:    false,
			expectedOutput:   []string{"No tag violations found"},
		},
		{
			name:             "json",
			filePathOrDir:    "testdata/cloudformation/single_resource.json",
			cliArgs:          []string{"--tags", "Owner,Environment"},
			expectedExitCode: 0,
			expectedError:    false,
			expectedOutput:   []string{"No tag violations found"},
		},
		{
			name:             "yaml missing tags",
			filePathOrDir:    "testdata/cloudformation/single_resource.yml",
			cliArgs:          []string{"--tags", "Owner,Environment,Project"},
			expectedExitCode: 1,
			expectedError:    true,
			expectedOutput:   []string{`AWS::S3::Bucket "this"`, "Missing tags: Project"},
		},
		{
			name:             "json missing tags",
			filePathOrDir:    "testdata/cloudformation/single_resource.json",
			cliArgs:          []string{"--tags", "Owner,Environment,Project"},
			expectedExitCode: 1,
			expectedError:    true,
			expectedOutput:   []string{`AWS::S3::Bucket "this"`, "Missing tags: Project"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			argsForRun := append([]string{tc.filePathOrDir}, tc.cliArgs...)
			output, err, exitCode := runTagNag(t, argsForRun...)

			if tc.expectedError && err == nil {
				t.Errorf("Expected an error from command execution, but got none. Output:\n%s", output)
			}
			if !tc.expectedError && err != nil {
				t.Errorf("Expected no error from command execution, but got: %v. Output:\n%s", err, output)
			}

			if exitCode != tc.expectedExitCode {
				t.Errorf("Expected exit code %d, got %d. Output:\n%s", tc.expectedExitCode, exitCode, output)
			}

			for _, expectedStr := range tc.expectedOutput {
				if !strings.Contains(output, expectedStr) {
					t.Errorf("Output missing expected string '%s'. Output:\n%s", expectedStr, output)
				}
			}
		})
	}
}
