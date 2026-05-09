package output

import (
	"fmt"
	"log"
	"os"

	"github.com/jakebark/tag-nag/internal/shared"
)

// ProcessOutput handles the output formatting and exit logic
func ProcessOutput(violations []shared.Violation, format shared.OutputFormat, dryRun bool, outputFile string, warning bool) {
	formatter := GetFormatter(format)
	formattedOutput, err := formatter.Format(violations)
	if err != nil {
		log.Fatalf("Error formatting output: %v", err)
	}

	if len(formattedOutput) > 0 {
		if outputFile != "" {
			if err := os.WriteFile(outputFile, formattedOutput, 0644); err != nil {
				log.Fatalf("Error writing output file: %v", err)
			}
			log.Printf("Output written to %s", outputFile)
		} else {
			fmt.Print(string(formattedOutput))
		}
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
	} else if nonSkippedCount > 0 && warning {
		log.Printf("\033[33mFound %d tag violation(s)\033[0m\n", nonSkippedCount)
		os.Exit(2)
	} else if nonSkippedCount > 0 {
		log.Printf("\033[31mFound %d tag violation(s)\033[0m\n", nonSkippedCount)
		os.Exit(1)
	} else {
		log.Println("No tag violations found")
		os.Exit(0)
	}
}

