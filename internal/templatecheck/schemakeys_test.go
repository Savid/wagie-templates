package templatecheck

import (
	"strings"
	"testing"
)

func TestStraySchemaKeysFlagsCommaSplitDescription(t *testing.T) {
	source := SchemaSource{
		Path: "bad.yaml",
		Data: []byte(`outputs:
  values:
    metric_limit:
      required: true
      schema: { type: number, description: the theoretical ceiling, named to bind verbatim }
`),
	}

	issues, err := StraySchemaKeys([]SchemaSource{source})
	if err != nil {
		t.Fatalf("StraySchemaKeys: %v", err)
	}

	if len(issues) != 1 {
		t.Fatalf("expected 1 issue, got %d: %v", len(issues), issues)
	}

	if issues[0].Key != "named to bind verbatim" {
		t.Fatalf("unexpected key %q", issues[0].Key)
	}

	if !strings.Contains(issues[0].Error(), "bad.yaml:5") {
		t.Fatalf("expected file:line in error, got %q", issues[0].Error())
	}
}

func TestStraySchemaKeysAcceptsVocabularyAndNestedSchemas(t *testing.T) {
	source := SchemaSource{
		Path: "good.yaml",
		Data: []byte(`outputs:
  values:
    trials:
      schema: &trials
        type: array
        minItems: 1
        items:
          type: object
          properties:
            round: { type: number }
            status: { type: string, enum: [pass, fail] }
            note: { type: string, description: "quoted, comma-safe description" }
          required: [round, status]
    echo:
      schema: *trials
`),
	}

	issues, err := StraySchemaKeys([]SchemaSource{source})
	if err != nil {
		t.Fatalf("StraySchemaKeys: %v", err)
	}

	if len(issues) != 0 {
		t.Fatalf("expected no issues, got %v", issues)
	}
}

func TestStraySchemaKeysFlagsSplitInsideNestedProperties(t *testing.T) {
	source := SchemaSource{
		Path: "nested.yaml",
		Data: []byte(`outputs:
  values:
    trials:
      schema:
        type: array
        items:
          type: object
          properties:
            hypothesis: { type: string, description: what was tried, one line }
`),
	}

	issues, err := StraySchemaKeys([]SchemaSource{source})
	if err != nil {
		t.Fatalf("StraySchemaKeys: %v", err)
	}

	if len(issues) != 1 || issues[0].Key != "one line" {
		t.Fatalf("expected the nested split key, got %v", issues)
	}
}
