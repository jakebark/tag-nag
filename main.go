package main

import (
	"fmt"
	"tag_nag/internal/inputs"
	"tag_nag/internal/terraform"
)

func main() {
	userInput := inputs.ParseFlags()
	fmt.Printf("🔍 Scanning: %s\n", userInput.Directory)

	terraform.ScanDirectory(userInput.Directory, userInput.RequiredTags, userInput.CaseInsensitive)
}
