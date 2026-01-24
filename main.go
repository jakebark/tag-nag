package main

import (
	"fmt"
	"log"
	"os"

	"github.com/jakebark/tag-nag/internal/cloudformation"
	"github.com/jakebark/tag-nag/internal/inputs"
	"github.com/jakebark/tag-nag/internal/output"
	"github.com/jakebark/tag-nag/internal/shared"
	"github.com/jakebark/tag-nag/internal/terraform"
)

func main() {
	log.SetFlags(0) // remove timestamp from prints

	userInput := inputs.ParseFlags()

	if userInput.DryRun {
		log.Printf("\033[32mDry-run: %s\033[0m\n", userInput.Directory)
	} else {
		log.Printf("\033[33mScanning: %s\033[0m\n", userInput.Directory)
	}

	tfViolations := terraform.ProcessDirectory(userInput.Directory, userInput.RequiredTags, userInput.CaseInsensitive, userInput.Skip)
	cfnViolations := cloudformation.ProcessDirectory(userInput.Directory, userInput.RequiredTags, userInput.CaseInsensitive, userInput.CfnSpecPath, userInput.Skip)

	var allViolations []shared.Violation
	allViolations = append(allViolations, tfViolations...)
	allViolations = append(allViolations, cfnViolations...)

	formatter := output.GetFormatter(userInput.OutputFormat)
	formattedOutput, err := formatter.Format(allViolations)
	if err != nil {
		log.Fatalf("Error formatting output: %v", err)
	}

	if len(formattedOutput) > 0 {
		fmt.Print(string(formattedOutput))
	}

	nonSkippedCount := 0
	for _, v := range allViolations {
		if !v.Skip {
			nonSkippedCount++
		}
	}

	if nonSkippedCount > 0 && userInput.DryRun {
		log.Printf("\033[32mFound %d tag violation(s)\033[0m\n", nonSkippedCount)
		os.Exit(0)
	} else if nonSkippedCount > 0 {
		log.Printf("\033[31mFound %d tag violation(s)\033[0m\n", nonSkippedCount)
		os.Exit(1)
	} else {
		log.Println("No tag violations found")
		os.Exit(0)
	}
}
