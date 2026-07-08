package templatecheck

import (
	"strings"
	"testing"
)

const (
	driftPathA = "ethereum/a.yaml"
	driftPathB = "ethereum/b.yaml"
)

const evidenceItemYAML = `
kind: Template
tasks:
  collect:
    outputs:
      values:
        evidence:
          schema:
            type: object
            properties:
              source: { type: string }
              ref: { type: string }
              at: { type: string }
              detail: { type: string, description: what was observed }
            required: [source, ref, detail]
`

func TestSchemaDriftFlagsStructuralDifference(t *testing.T) {
	t.Parallel()

	drifted := strings.Replace(
		evidenceItemYAML,
		"required: [source, ref, detail]",
		"required: [source, ref]",
		1,
	)

	issues, err := SchemaDrift([]SchemaSource{
		{Path: driftPathA, Data: []byte(evidenceItemYAML)},
		{Path: driftPathB, Data: []byte(drifted)},
	})
	if err != nil {
		t.Fatalf("SchemaDrift returned error: %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("expected one drift issue, got %d: %#v", len(issues), issues)
	}
	if issues[0].Shape != "evidence item" {
		t.Fatalf("unexpected shape: %q", issues[0].Shape)
	}
	if len(issues[0].Variants) != 2 {
		t.Fatalf("expected two variants, got %d", len(issues[0].Variants))
	}
	for _, variant := range issues[0].Variants {
		if len(variant.Sites) != 1 {
			t.Fatalf("expected one site per variant, got %#v", variant.Sites)
		}
	}
}

func TestSchemaDriftIgnoresDescriptionDifferences(t *testing.T) {
	t.Parallel()

	reworded := strings.Replace(
		evidenceItemYAML,
		"description: what was observed",
		"description: verbatim values preserved",
		1,
	)

	issues, err := SchemaDrift([]SchemaSource{
		{Path: driftPathA, Data: []byte(evidenceItemYAML)},
		{Path: driftPathB, Data: []byte(reworded)},
	})
	if err != nil {
		t.Fatalf("SchemaDrift returned error: %v", err)
	}
	if len(issues) != 0 {
		t.Fatalf("expected no drift for description-only difference, got %#v", issues)
	}
}

func TestSchemaDriftFollowsAliases(t *testing.T) {
	t.Parallel()

	anchored := `
kind: Template
tasks:
  a:
    outputs:
      values:
        evidence:
          schema: &evidence_item_schema
            type: object
            properties:
              source: { type: string }
              ref: { type: string }
              at: { type: string }
              detail: { type: string }
            required: [source, ref, detail]
  b:
    outputs:
      values:
        evidence:
          schema: *evidence_item_schema
`

	issues, err := SchemaDrift([]SchemaSource{
		{Path: driftPathA, Data: []byte(anchored)},
		{Path: driftPathB, Data: []byte(evidenceItemYAML)},
	})
	if err != nil {
		t.Fatalf("SchemaDrift returned error: %v", err)
	}
	if len(issues) != 0 {
		t.Fatalf("expected aliased and inline declarations to match, got %#v", issues)
	}
}
