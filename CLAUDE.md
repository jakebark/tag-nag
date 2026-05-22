# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

tag-nag is a Go CLI tool that validates AWS tags in Terraform (.tf) and CloudFormation (.yaml/.json) files. It checks resources for required tag keys and optionally validates tag values.

## Build & Test Commands

```bash
go build -o tag-nag                          # Build binary
go test ./...                                # Run all tests
go test -v ./...                             # Verbose test output
go test -run TestTerraform ./                # Run specific integration test
go test ./internal/terraform                 # Run tests for a single package
go test -run TestParseTags ./internal/inputs  # Run single test in a package
```

There is no separate lint command configured. The CI matrix tests against Go 1.21–1.24 on Ubuntu and macOS.

## Architecture

**Entry point:** `main.go` orchestrates a pipeline: parse input → scan Terraform → scan CloudFormation → combine violations → format output → exit with code 0 (clean/dry-run) or 1 (violations found).

**Core packages under `internal/`:**

- **inputs/** — CLI flag parsing (pflag) and `.tag-nag.yml` config file loading. Produces a `UserInput` struct consumed by all other packages.
- **terraform/** — Walks directories for `.tf` files, parses HCL with `hashicorp/hcl/v2`, resolves variables/locals, handles provider `default_tags`, and checks each resource's tags against requirements.
- **cloudformation/** — Walks directories for YAML/JSON files, parses templates with `yaml.v3`, optionally filters taggable resource types via a CloudFormation spec file (`--cfn-spec`), and checks resource tags.
- **output/** — Formatter factory (`text`, `json`, `junit`) selected by `--output` flag. Each formatter implements a `Format([]Violation) string` interface.
- **shared/** — Shared types (`Violation`, `TagMap`, `OutputFormat`) and helpers for tag filtering and case normalization.
- **config/** — Constants like ignore-comment markers and stdlib function definitions for HCL evaluation.

**Key data flow:** `UserInput.RequiredTags` is a `map[string][]string` where keys are tag names and values are optional allowed values (e.g., `{"Environment": ["Dev", "Prod"]}`). Both Terraform and CloudFormation processors return `[]Violation` which are merged and passed to the output formatter.

**Test fixtures** live in `testdata/` with subdirectories for terraform, cloudformation, and config file test cases. Integration tests in `main_test.go` build the actual binary and run CLI-level tests.
