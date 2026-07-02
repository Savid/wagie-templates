package templatecheck

import "testing"

func TestConditionalNeedsFindsDependedOnIfTask(t *testing.T) {
	t.Parallel()

	issues, err := ConditionalNeeds([]byte(`
kind: Template
tasks:
  optional:
    if: "${{ false }}"
    instructions: optional
  dependent:
    needs: [optional]
    instructions: dependent
`))
	if err != nil {
		t.Fatalf("ConditionalNeeds returned error: %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("expected one issue, got %d: %#v", len(issues), issues)
	}
	if issues[0].Task != "optional" {
		t.Fatalf("unexpected task: %q", issues[0].Task)
	}
	if len(issues[0].Dependents) != 1 || issues[0].Dependents[0] != "dependent" {
		t.Fatalf("unexpected dependents: %#v", issues[0].Dependents)
	}
	if issues[0].Line == 0 {
		t.Fatal("expected source line to be recorded")
	}
}

func TestConditionalNeedsAllowsLeafConditional(t *testing.T) {
	t.Parallel()

	issues, err := ConditionalNeeds([]byte(`
kind: Template
tasks:
  cleanup:
    if: "${{ inputs.values.cleanup }}"
    instructions: cleanup
  done:
    instructions: done
`))
	if err != nil {
		t.Fatalf("ConditionalNeeds returned error: %v", err)
	}
	if len(issues) != 0 {
		t.Fatalf("expected no issues, got %#v", issues)
	}
}
