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
	binaryName                = "tag-nag"
	requiredTagsCLI           = "Owner,Environment,Project"
	requiredTagsWithValuesCLI = "Owner,Environment[dev,prod],Project"
)

func TestMain(m *testing.M) {
	fmt.Println("Building tag-nag binary for E2E tests...")
	cmd := exec.Command("go", "build", "-o", binaryName)
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to build %s: %v\n", binaryName, err)
		os.Exit(1)
	}

	exitVal := m.Run()
	os.Remove(binaryName)
	os.Exit(exitVal)
}

func runTagNagCommand(t *testing.T, args ...string) (string, error, int) {
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

func TestTerraform_PassBasic(t *testing.T) {
	output, err, exitCode := runTagNagCommand(t, "testdata/terraform/pass_basic", "--tags", requiredTagsCLI)
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

func TestTerraform_FailBasic(t *testing.T) {
	output, err, exitCode := runTagNagCommand(t, "testdata/terraform/fail_basic", "--tags", requiredTagsCLI)
	if err == nil {
		t.Errorf("Expected an error due to violations, but got none. Output:\n%s", output)
	}
	if exitCode != 1 {
		t.Errorf("Expected exit code 1, got %d. Output:\n%s", exitCode, output)
	}
	if !strings.Contains(output, `aws_s3_bucket "this"`) || !strings.Contains(output, "Missing tags: Environment, Project") {
		t.Errorf("Output missing expected violation details for fail_basic. Output:\n%s", output)
	}
}
