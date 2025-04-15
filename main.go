package main

import (
	"fmt"
	"os"

	"github.com/jakebark/tag-nag/internal/cloudformation"
	"github.com/jakebark/tag-nag/internal/inputs"
	"github.com/jakebark/tag-nag/internal/shared"
	"github.com/jakebark/tag-nag/internal/terraform"
)

func main() {
	userInput := inputs.ParseFlags()

	if userInput.DryRun {
		fmt.Printf("\033[32mDry-run: %s\033[0m\n", userInput.Directory)
	} else {
		fmt.Printf("\033[33mScanning: %s\033[0m\n", userInput.Directory)
	}

	tfViolations := terraform.ProcessDirectory(userInput.Directory, userInput.RequiredTags, userInput.CaseInsensitive)
	cfnViolations := cloudformation.ProcessDirectory(userInput.Directory, userInput.RequiredTags, userInput.CaseInsensitive)

	violations := append(tfViolations, cfnViolations...)

	violationsByFile := make(map[string][]shared.Violation)
	for _, v := range violations {
		violationsByFile[v.FilePath] = append(violationsByFile[v.FilePath], v)
	}

	for filePath, fileViolations := range violationsByFile {
		shared.PrintViolations(filePath, fileViolations)
	}

	violationCount := 0
	for _, v := range violations {
		if !v.Skip {
			violationCount++
		}
	}

	if violationCount > 0 && userInput.DryRun {
		fmt.Printf("\033[32mFound %d tag violation(s)\033[0m\n", violations)
		os.Exit(0)
	} else if violationCount > 0 {
		fmt.Printf("\033[31mFound %d tag violation(s)\033[0m\n", violations)
		os.Exit(1)
	} else {
		fmt.Println("No tag violations found")
		os.Exit(0)
	}
}
