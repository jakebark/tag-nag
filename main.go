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
	fmt.Printf("Scanning: %s\n", userInput.Directory)

	tfViolations := terraform.ProcessDirectory(userInput.Directory, userInput.RequiredTags, userInput.CaseInsensitive, userInput.ShowSkips)
	cfnViolations := cloudformation.ProcessDirectory(userInput.Directory, userInput.RequiredTags, userInput.CaseInsensitive, userInput.ShowSkips)

	violations := tfViolations + cfnViolations

	if violations > 0 {
		fmt.Printf("\033[31mFound %d tag violation(s)\033[0m\n", violations)
		os.Exit(1)
	}
	fmt.Println("No tag violations found")

}
