package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
)

var (
	binaryName       = "tag-nag"
	tags             = "Owner,Environment"
	tagsMissing      = "Owner,Environment,Project"
	tagsLower        = "owner,environment"
	tagValues        = "Owner,Environment[dev,prod]"
	tagValuesMissing = "Owner,Environment[test]"
)

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
	cmd := exec.Command("./"+binaryName, args...)
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

func TestTerraformPassSingleResource(t *testing.T) {
	output, err, exitCode := runTagNag(t, "testdata/terraform/single_resource.tf", "--tags", tags)
	if err != nil {
		t.Errorf("Expected no error, got exit code %d, err: %v, output:\n%s", exitCode, err, output)
	}
	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d. Output:\n%s", exitCode, output)
	}
	if !strings.Contains(output, "No tag violations found") {
		t.Errorf("Expected 'No tag violations found', output:\n%s", output)
	}
}

func TestTerraformFailSingleResource(t *testing.T) {
	output, err, exitCode := runTagNag(t, "testdata/terraform/single_resource.tf", "--tags", tagsMissing)
	if err == nil {
		t.Errorf("Expected an error due to violations, but got none. Output:\n%s", output)
	}
	if exitCode != 1 {
		t.Errorf("Expected exit code 1, got %d. Output:\n%s", exitCode, output)
	}
	if !strings.Contains(output, `aws_s3_bucket "this"`) || !strings.Contains(output, "Missing tags: Project") {
		t.Errorf("Output missing expected violation details for fail_basic. Output:\n%s", output)
	}
}

func TestTerraformFailNoTags(t *testing.T) {
	output, err, exitCode := runTagNag(t, "testdata/terraform/no_tags.tf", "--tags", tags)
	if err == nil {
		t.Errorf("Expected an error due to violations, but got none. Output:\n%s", output)
	}
	if exitCode != 1 {
		t.Errorf("Expected exit code 1, got %d. Output:\n%s", exitCode, output)
	}
	if !strings.Contains(output, `aws_s3_bucket "this"`) || !strings.Contains(output, "Missing tags: Owner, Environment") {
		t.Errorf("Output missing expected violation details for fail_basic. Output:\n%s", output)
	}
}

func TestTerraformPassCaseInsensitive(t *testing.T) {
	output, err, exitCode := runTagNag(t, "testdata/terraform/single_resource.tf", "--tags", tagsLower, "-c")
	if err != nil {
		t.Errorf("Expected no error with case-insensitive, got exit code %d, err: %v, output:\n%s", exitCode, err, output)
	}
	if exitCode != 0 {
		t.Errorf("Expected exit code 0 with case-insensitive, got %d. Output:\n%s", exitCode, output)
	}
}

func TestTerraformFailCaseInsensitive(t *testing.T) {
	output, err, exitCode := runTagNag(t, "testdata/terraform/no_tags.tf", "--tags", tags)
	if err == nil {
		t.Errorf("Expected an error due to violations, but got none. Output:\n%s", output)
	}
	if exitCode != 1 {
		t.Errorf("Expected exit code 1, got %d. Output:\n%s", exitCode, output)
	}
	if !strings.Contains(output, `aws_s3_bucket "this"`) || !strings.Contains(output, "Missing tags: Owner, Environment") {
		t.Errorf("Output missing expected violation details for fail_basic. Output:\n%s", output)
	}
}
