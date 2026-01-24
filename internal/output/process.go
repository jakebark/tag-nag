package output

import (
	"fmt"
	"log"
	"os"

	"github.com/jakebark/tag-nag/internal/shared"
)

// ProcessOuput handles the output formatting and exit logic
func ProcessOutput(violations []shared.Violation, format shared.OutputFormat, dryRun bool) {
	formatter := GetFormatter(format)
	formattedOutput, err := formatter.Format(violations)
	if err != nil {
		log.Fatalf("Error formatting output: %v", err)
	}

	if len(formattedOutput) > 0 {
		fmt.Print(string(formattedOutput))
	}

	nonSkippedCount := 0
	for _, v := range violations {
		if !v.Skip {
			nonSkippedCount++
		}
	}

	if nonSkippedCount > 0 && dryRun {
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

