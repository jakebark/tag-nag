package main

import (
	"fmt"
	"os"

	"github.com/jakebark/tag-nag/internal/cloudformation"
	"github.com/jakebark/tag-nag/internal/inputs"
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

	violations := tfViolations + cfnViolations

	if violations > 0 && userInput.DryRun {
		fmt.Printf("\033[32mFound %d tag violation(s)\033[0m\n", violations)
		os.Exit(0)
	} else if violations > 0 {
		fmt.Printf("\033[31mFound %d tag violation(s)\033[0m\n", violations)
		os.Exit(1)
	} else {
		fmt.Println("No tag violations found")
		os.Exit(0)
	}

}
