// Command validate checks the wagie-templates library against the Wagie core
// templates using Wagie's executable preflight path — the same validation the
// combined-library test runs, exposed as a CLI for fast local iteration.
//
// Usage:
//
//	go run ./cmd/validate              # validate every family, report all
//	go run ./cmd/validate code         # report only files whose path contains "code"
//	go run ./cmd/validate code/code-reviewer.yaml research
//
// Filters are path substrings. The whole library (core + every family) is
// always loaded so that uses: composition resolves; filters only narrow what
// is reported.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/savid/wagie"
	"github.com/savid/wagie-templates/internal/templatecheck"
)

// families are the template directories owned by this repo.
var families = []string{"ethereum", "code", "research", "experiments"}

func main() {
	filters := os.Args[1:]

	coreFiles, err := wagie.CoreTemplateFiles()
	if err != nil {
		fmt.Fprintf(os.Stderr, "load core templates: %v\n", err)
		os.Exit(1)
	}

	files := make([]wagie.TemplateFile, 0, len(coreFiles)+64)
	files = append(files, coreFiles...)

	for _, dir := range families {
		loaded, loadErr := wagie.LoadTemplateFilesRecursive(dir)
		if loadErr != nil {
			fmt.Fprintf(os.Stderr, "load %s: %v\n", dir, loadErr)
			os.Exit(1)
		}
		for _, file := range loaded {
			file.Source = "wagie-templates"
			files = append(files, file)
		}
	}

	results, err := wagie.ValidateTemplateFiles(
		context.Background(),
		slog.New(slog.DiscardHandler),
		files,
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "validate: %v\n", err)
		os.Exit(1)
	}

	warnings := collectWarnings(files)
	reported, failed := report(results, warnings, filters)
	failed += reportSchemaDrift(files)
	fmt.Printf("\n%d reported, %d failed\n", reported, failed)
	if failed > 0 {
		os.Exit(1)
	}
}

// reportSchemaDrift compares the shared runbook-owned schema shapes across
// every family file (filters do not apply — drift is a cross-file property)
// and returns the number of drifted shapes.
func reportSchemaDrift(files []wagie.TemplateFile) int {
	sources := make([]templatecheck.SchemaSource, 0, len(files))
	for _, file := range files {
		if isCore(file.Path) {
			continue
		}
		sources = append(sources, templatecheck.SchemaSource{Path: relPath(file.Path), Data: file.Data})
	}

	issues, err := templatecheck.SchemaDrift(sources)
	if err != nil {
		fmt.Fprintf(os.Stderr, "schema drift check: %v\n", err)
		os.Exit(1)
	}
	for _, issue := range issues {
		fmt.Printf("FAIL  schema drift\n        - %s\n", issue.Error())
	}
	return len(issues)
}

// report prints one line per reported family file and returns the reported and
// failed counts.
func report(results []wagie.TemplateValidationResult, warnings map[string][]templateWarning, filters []string) (int, int) {
	reported, failed := 0, 0
	for _, result := range results {
		if isCore(result.Path) || !matchesFilter(result.Path, filters) {
			continue
		}
		reported++

		header := relPath(result.Path)
		if result.Name != "" {
			header = fmt.Sprintf("%s (%s@%s)", relPath(result.Path), result.Name, result.Version)
		}

		if result.Valid {
			fmt.Printf("ok    %s\n", header)
			printWarnings(warnings[result.Path])
			continue
		}

		failed++
		fmt.Printf("FAIL  %s\n", header)
		printErrors(result.Errors)
		printWarnings(warnings[result.Path])
	}
	return reported, failed
}

func printErrors(errs []wagie.TemplateValidationError) {
	for _, e := range errs {
		if e.Line > 0 {
			fmt.Printf("        - [%s] line %d: %s\n", e.Type, e.Line, e.Message)
			continue
		}
		fmt.Printf("        - [%s] %s\n", e.Type, e.Message)
	}
}

type templateWarning struct {
	Line    int
	Message string
}

func collectWarnings(files []wagie.TemplateFile) map[string][]templateWarning {
	out := map[string][]templateWarning{}
	for _, file := range files {
		if isCore(file.Path) {
			continue
		}
		issues, err := templatecheck.ConditionalNeeds(file.Data)
		if err != nil {
			out[file.Path] = append(out[file.Path], templateWarning{
				Message: fmt.Sprintf("topology warning skipped: %v", err),
			})
			continue
		}
		for _, issue := range issues {
			out[file.Path] = append(out[file.Path], templateWarning{
				Line:    issue.Line,
				Message: issue.Error(),
			})
		}
	}
	return out
}

func printWarnings(warnings []templateWarning) {
	for _, warning := range warnings {
		if warning.Line > 0 {
			fmt.Printf("        - [warning] line %d: %s\n", warning.Line, warning.Message)
			continue
		}
		fmt.Printf("        - [warning] %s\n", warning.Message)
	}
}

// isCore reports whether a result path belongs to the embedded Wagie core
// library rather than a repo family. CoreTemplateFiles sets Path to
// "templates/<name>"; family files carry an absolute "/<family>/" segment.
func isCore(path string) bool {
	for _, dir := range families {
		if strings.Contains(path, "/"+dir+"/") {
			return false
		}
	}
	return true
}

// matchesFilter reports whether a family file path matches any path-substring
// filter. With no filters, everything matches.
func matchesFilter(path string, filters []string) bool {
	if len(filters) == 0 {
		return true
	}
	rel := relPath(path)
	for _, f := range filters {
		if strings.Contains(rel, f) {
			return true
		}
	}
	return false
}

func relPath(path string) string {
	for _, dir := range families {
		if idx := strings.Index(path, "/"+dir+"/"); idx >= 0 {
			return path[idx+1:]
		}
	}
	if idx := strings.Index(path, "templates/"); idx >= 0 {
		return path[idx:]
	}
	return path
}
