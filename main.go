package main

import (
	"log"

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

	output.ProcessOutput(allViolations, userInput.OutputFormat, userInput.DryRun)
}
