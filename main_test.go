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

func TestInputs(t *testing.T) {
	testCases := []testCases{
		{
			name:             "no dir",
			filePathOrDir:    "",
			cliArgs:          []string{"--tags", "Owner"},
			expectedExitCode: 1,
			expectedError:    true,
			expectedOutput:   []string{"Error: Please specify a directory or file to scan."},
		},
		{
			name:             "no tags",
			filePathOrDir:    "testdata/terraform/tags.tf",
			cliArgs:          []string{},
			expectedExitCode: 1,
			expectedError:    true,
			expectedOutput:   []string{"Error: Please specify required tags using --tags"},
		},
		{
			name:             "dry run",
			filePathOrDir:    "testdata/terraform/tags.tf",
			cliArgs:          []string{"--tags", "Owner,Environment,Project", "--dry-run"},
			expectedExitCode: 0,
			expectedError:    false,
			expectedOutput:   []string{"Dry-run:", `aws_s3_bucket "this"`, "Missing tags: Project"},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var argsForRun []string
			if tc.filePathOrDir != "" {
				argsForRun = append(argsForRun, tc.filePathOrDir)
			}
			argsForRun = append(argsForRun, tc.cliArgs...)

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

func TestTerraform(t *testing.T) {
	testCases := []testCases{
		{
			name:             "tags",
			filePathOrDir:    "testdata/terraform/tags.tf",
			cliArgs:          []string{"--tags", "Owner,Environment"},
			expectedExitCode: 0,
			expectedError:    false,
			expectedOutput:   []string{"No tag violations found"},
		},
		{
			name:             "missing tags",
			filePathOrDir:    "testdata/terraform/tags.tf",
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
			filePathOrDir:    "testdata/terraform/tags.tf",
			cliArgs:          []string{"--tags", "owner,environment", "-c"},
			expectedExitCode: 0,
			expectedError:    false,
			expectedOutput:   []string{"No tag violations found"},
		},
		{
			name:             "lower case",
			filePathOrDir:    "testdata/terraform/tags.tf",
			cliArgs:          []string{"--tags", "owner,environment"},
			expectedExitCode: 1,
			expectedError:    true,
			expectedOutput:   []string{`aws_s3_bucket "this"`, "Missing tags: owner, environment"},
		},
		{
			name:             "tag values",
			filePathOrDir:    "testdata/terraform/tags.tf",
			cliArgs:          []string{"--tags", "Owner,Environment[dev,prod]"},
			expectedExitCode: 0,
			expectedError:    false,
			expectedOutput:   []string{"No tag violations found"},
		},
		{
			name:             "missing tag value",
			filePathOrDir:    "testdata/terraform/tags.tf",
			cliArgs:          []string{"--tags", "Owner,Environment[test]"},
			expectedExitCode: 1,
			expectedError:    true,
			expectedOutput:   []string{`aws_s3_bucket "this"`, "Missing tags: Environment[test]"},
		},
		{
			name:             "tag values case insensitive",
			filePathOrDir:    "testdata/terraform/tags.tf",
			cliArgs:          []string{"--tags", "Owner,Environment[Dev,Prod]", "-c"},
			expectedExitCode: 0,
			expectedError:    false,
			expectedOutput:   []string{"No tag violations found"},
		},
		{
			name:             "provider",
			filePathOrDir:    "testdata/terraform/provider.tf",
			cliArgs:          []string{"--tags", "Owner,Environment,Project,Source"},
			expectedExitCode: 0,
			expectedError:    false,
			expectedOutput:   []string{"Found Terraform default tags for provider aws", "No tag violations found"},
		},
		{
			name:             "provider",
			filePathOrDir:    "testdata/terraform/provider.tf",
			cliArgs:          []string{"--tags", "Owner,Environment,Project,Source"},
			expectedExitCode: 0,
			expectedError:    false,
			expectedOutput:   []string{"Found Terraform default tags for provider aws", "No tag violations found"},
		},
		{
			name:             "provider case insensitive",
			filePathOrDir:    "testdata/terraform/provider.tf",
			cliArgs:          []string{"--tags", "owner,environment,project,source", "-c"},
			expectedExitCode: 0,
			expectedError:    false,
			expectedOutput:   []string{"Found Terraform default tags for provider aws", "No tag violations found"},
		},
		{
			name:             "provider tag values",
			filePathOrDir:    "testdata/terraform/provider.tf",
			cliArgs:          []string{"--tags", "Owner,Environment[dev,prod],Project,Source[my-repo]"},
			expectedExitCode: 0,
			expectedError:    false,
			expectedOutput:   []string{"Found Terraform default tags for provider aws", "No tag violations found"},
		},
		{
			name:             "variable tags",
			filePathOrDir:    "testdata/terraform/referenced_tags.tf",
			cliArgs:          []string{"--tags", "Owner,Environment"},
			expectedExitCode: 0,
			expectedError:    false,
			expectedOutput:   []string{"No tag violations found"},
		},
		{
			name:             "variable value",
			filePathOrDir:    "testdata/terraform/referenced_values.tf",
			cliArgs:          []string{"--tags", "Owner[jakebark],Environment"},
			expectedExitCode: 0,
			expectedError:    false,
			expectedOutput:   []string{"No tag violations found"},
		},
		{
			name:             "local value",
			filePathOrDir:    "testdata/terraform/referenced_values.tf",
			cliArgs:          []string{"--tags", "Owner,Environment[dev,prod]"},
			expectedExitCode: 0,
			expectedError:    false,
			expectedOutput:   []string{"No tag violations found"},
		},
		{
			name:             "variable value case insensitive",
			filePathOrDir:    "testdata/terraform/referenced_values.tf",
			cliArgs:          []string{"--tags", "Owner[Jakebark],Environment", "-c"},
			expectedExitCode: 0,
			expectedError:    false,
			expectedOutput:   []string{"No tag violations found"},
		},
		{
			name:             "local value case insensitive",
			filePathOrDir:    "testdata/terraform/referenced_values.tf",
			cliArgs:          []string{"--tags", "Owner,Environment[DEV,PROD]", "-c"},
			expectedExitCode: 0,
			expectedError:    false,
			expectedOutput:   []string{"No tag violations found"},
		},
		{
			name:             "interpolation",
			filePathOrDir:    "testdata/terraform/referenced_values.tf",
			cliArgs:          []string{"--tags", "Owner,Environment,Project[112233],Source[my-repo]"},
			expectedExitCode: 0,
			expectedError:    false,
			expectedOutput:   []string{"No tag violations found"},
		},
		{
			name:             "interpolation missing value",
			filePathOrDir:    "testdata/terraform/referenced_values.tf",
			cliArgs:          []string{"--tags", "Owner,Environment,Project[112233],Source[not-my-repo]"},
			expectedExitCode: 1,
			expectedError:    true,
			expectedOutput:   []string{`aws_s3_bucket "this"`, "Missing tags: Source[not-my-repo]"},
		},
		{
			name:             "example repo",
			filePathOrDir:    "testdata/terraform/example_repo",
			cliArgs:          []string{"--tags", "Owner,Environment"},
			expectedExitCode: 1,
			expectedError:    true,
			expectedOutput:   []string{"Found Terraform default tags for provider aws", `aws_s3_bucket "baz"`, "Found 1 tag violation(s)"},
		},
		{
			name:             "ignore",
			filePathOrDir:    "testdata/terraform/ignore.tf",
			cliArgs:          []string{"--tags", "Owner,Environment,Project"},
			expectedExitCode: 0,
			expectedError:    false,
			expectedOutput:   []string{`aws_s3_bucket "this" skipped`},
		},
		{
			name:             "ignore all",
			filePathOrDir:    "testdata/terraform/ignore_all.tf",
			cliArgs:          []string{"--tags", "Owner,Environment,Project"},
			expectedExitCode: 0,
			expectedError:    false,
			expectedOutput:   []string{`aws_s3_bucket "this" skipped`},
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

func TestCloudFormation(t *testing.T) {
	testCases := []testCases{
		{
			name:             "yml",
			filePathOrDir:    "testdata/cloudformation/tags.yml",
			cliArgs:          []string{"--tags", "Owner,Environment"},
			expectedExitCode: 0,
			expectedError:    false,
			expectedOutput:   []string{"No tag violations found"},
		},
		{
			name:             "yaml",
			filePathOrDir:    "testdata/cloudformation/tags.yaml",
			cliArgs:          []string{"--tags", "Owner,Environment"},
			expectedExitCode: 0,
			expectedError:    false,
			expectedOutput:   []string{"No tag violations found"},
		},
		{
			name:             "json",
			filePathOrDir:    "testdata/cloudformation/tags.json",
			cliArgs:          []string{"--tags", "Owner,Environment"},
			expectedExitCode: 0,
			expectedError:    false,
			expectedOutput:   []string{"No tag violations found"},
		},
		{
			name:             "yaml missing tags",
			filePathOrDir:    "testdata/cloudformation/tags.yml",
			cliArgs:          []string{"--tags", "Owner,Environment,Project"},
			expectedExitCode: 1,
			expectedError:    true,
			expectedOutput:   []string{`AWS::S3::Bucket "this"`, "Missing tags: Project"},
		},
		{
			name:             "json missing tags",
			filePathOrDir:    "testdata/cloudformation/tags.json",
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
