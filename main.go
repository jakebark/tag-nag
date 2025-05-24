package main

import (
	"log"
	"os"

	"github.com/jakebark/tag-nag/internal/cloudformation"
	"github.com/jakebark/tag-nag/internal/inputs"
	"github.com/jakebark/tag-nag/internal/terraform"
)

var version = "dev"

func main() {
	log.SetFlags(0) // remove timestamp from prints

	userInput := inputs.ParseFlags()

	if userInput.DryRun {
		log.Printf("\033[32mDry-run: %s\033[0m\n", userInput.Directory)
	} else {
		log.Printf("\033[33mScanning: %s\033[0m\n", userInput.Directory)
	}

	tfViolations := terraform.ProcessDirectory(userInput.Directory, userInput.RequiredTags, userInput.CaseInsensitive)
	cfnViolations := cloudformation.ProcessDirectory(userInput.Directory, userInput.RequiredTags, userInput.CaseInsensitive, userInput.CfnSpecPath)

	violations := tfViolations + cfnViolations

	if violations > 0 && userInput.DryRun {
		log.Printf("\033[32mFound %d tag violation(s)\033[0m\n", violations)
		os.Exit(0)
	} else if violations > 0 {
		log.Printf("\033[31mFound %d tag violation(s)\033[0m\n", violations)
		os.Exit(1)
	} else {
		log.Println("No tag violations found")
		os.Exit(0)
	}
}
