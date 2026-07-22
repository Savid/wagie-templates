package wagietemplates_test

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"testing"

	"github.com/savid/wagie"
)

const templateSource = "wagie-templates"

func TestTemplatesValidateAgainstCombinedLibrary(t *testing.T) {
	t.Parallel()

	coreFiles, err := wagie.CoreTemplateFiles()
	if err != nil {
		t.Fatalf("load core templates: %v", err)
	}

	dirs := []struct {
		path   string
		source string
	}{
		{path: "ethereum", source: templateSource},
		{path: "code", source: templateSource},
		{path: "research", source: templateSource},
		{path: "experiments", source: templateSource},
	}

	files := make([]wagie.TemplateFile, 0, len(coreFiles)+64)
	files = append(files, coreFiles...)

	for _, dir := range dirs {
		loaded, loadErr := wagie.LoadTemplateFilesRecursive(dir.path)
		if loadErr != nil {
			t.Fatalf("load %s: %v", dir.path, loadErr)
		}

		for _, file := range loaded {
			file.Source = dir.source
			files = append(files, file)
		}
	}

	results, err := wagie.ValidateTemplateFiles(
		context.Background(),
		slog.New(slog.DiscardHandler),
		files,
	)
	if err != nil {
		t.Fatalf("validate files: %v", err)
	}

	var failures []string
	for _, result := range results {
		if result.Valid {
			continue
		}

		lines := make([]string, 0, len(result.Errors)+1)
		header := result.Path
		if result.Name != "" {
			header = fmt.Sprintf("%s (%s@%s)", result.Path, result.Name, result.Version)
		}
		lines = append(lines, header)
		for _, err := range result.Errors {
			if err.Line > 0 {
				lines = append(lines, fmt.Sprintf("  - [%s] line %d: %s", err.Type, err.Line, err.Message))
				continue
			}
			lines = append(lines, fmt.Sprintf("  - [%s] %s", err.Type, err.Message))
		}
		failures = append(failures, strings.Join(lines, "\n"))
	}

	if len(failures) > 0 {
		t.Fatalf("template validation failures:\n%s", strings.Join(failures, "\n\n"))
	}
}
