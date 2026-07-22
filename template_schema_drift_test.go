package wagietemplates_test

import (
	"testing"

	"github.com/savid/wagie"
	"github.com/savid/wagie-templates/internal/templatecheck"
)

// TestSharedSchemasDoNotDrift enforces that the runbook-owned shapes copied
// between template files (the spec grammar has no cross-file schema import)
// stay structurally identical. Descriptions may vary per site.
func TestSharedSchemasDoNotDrift(t *testing.T) {
	t.Parallel()

	dirs := []string{"ethereum", "code", "research", "experiments"}
	sources := make([]templatecheck.SchemaSource, 0, 64)

	for _, dir := range dirs {
		loaded, err := wagie.LoadTemplateFilesRecursive(dir)
		if err != nil {
			t.Fatalf("load %s: %v", dir, err)
		}
		for _, file := range loaded {
			sources = append(sources, templatecheck.SchemaSource{Path: file.Path, Data: file.Data})
		}
	}

	issues, err := templatecheck.SchemaDrift(sources)
	if err != nil {
		t.Fatalf("schema drift check: %v", err)
	}
	for _, issue := range issues {
		t.Errorf("%s", issue.Error())
	}
}
