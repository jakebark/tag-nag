package main

import (
	"fmt"
	"os"

	"github.com/jakebark/tag-nag/internal/inputs"
	"github.com/jakebark/tag-nag/internal/terraform"
)

func main() {
	userInput := inputs.ParseFlags()
	fmt.Printf("🔍 Scanning: %s\n", userInput.Directory)

	violations := terraform.ProcessDirectory(userInput.Directory, userInput.RequiredTags, userInput.CaseInsensitive)

	if violations > 0 {
		fmt.Printf("Found %d tag violation(s)\n", violations)
		os.Exit(1)
	}
	fmt.Println("No tag violations found")
}
