package wagietemplates_test

import (
	"fmt"
	"os"
	"testing"

	"gopkg.in/yaml.v3"
)

// TestExperimentBridgeMirrorsExperimentLoop enforces the deliberate non-`uses:`
// bridge between ethereum/devnet-issue-to-experiment and
// experiments/experiment-loop: the bridge's `experiment_inputs` bundle promises
// to mirror experiment-loop's caller inputs by name, but nothing in the spec
// grammar validates a by-name mirror, so a rename in either file would
// silently strand the bundle.
func TestExperimentBridgeMirrorsExperimentLoop(t *testing.T) {
	t.Parallel()

	bridge := loadYAMLMap(t, "ethereum/devnet-issue-to-experiment.yaml")
	loop := loadYAMLMap(t, "experiments/experiment-loop.yaml")

	bundleSchema := dig(t, bridge, "outputs", "values", "experiment_inputs", "schema")
	bundleProps, ok := bundleSchema["properties"].(map[string]any)
	if !ok {
		t.Fatal("experiment_inputs schema has no properties map")
	}

	loopInputs := dig(t, loop, "inputs", "values")

	for name := range bundleProps {
		if _, exists := loopInputs[name]; !exists {
			t.Errorf("experiment_inputs.%s has no matching experiment-loop caller input", name)
		}
	}

	bundleRequired := stringSet(t, bundleSchema["required"])
	for name, raw := range loopInputs {
		input, isMapping := raw.(map[string]any)
		if !isMapping {
			t.Fatalf("experiment-loop input %s is not a mapping", name)
		}
		if required, _ := input["required"].(bool); required && !bundleRequired[name] {
			t.Errorf("experiment-loop requires %s but experiment_inputs does not carry it as required", name)
		}
	}

	bridgeEnum := enumValues(t, bundleProps["metric_direction"])
	loopEnum := enumValues(t, dig(t, loopInputs, "metric_direction")["schema"])
	if fmt.Sprint(bridgeEnum) != fmt.Sprint(loopEnum) {
		t.Errorf("metric_direction enum drifted: bridge %v vs experiment-loop %v", bridgeEnum, loopEnum)
	}
}

func loadYAMLMap(t *testing.T, path string) map[string]any {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}

	var out map[string]any
	if unmarshalErr := yaml.Unmarshal(data, &out); unmarshalErr != nil {
		t.Fatalf("parse %s: %v", path, unmarshalErr)
	}

	return out
}

// dig walks nested string-keyed mappings, failing the test on a missing or
// mistyped step so assertions read as one line at the call site.
func dig(t *testing.T, node map[string]any, path ...string) map[string]any {
	t.Helper()

	current := node
	for _, key := range path {
		next, ok := current[key].(map[string]any)
		if !ok {
			t.Fatalf("missing or non-mapping key %q while digging %v", key, path)
		}
		current = next
	}

	return current
}

func stringSet(t *testing.T, raw any) map[string]bool {
	t.Helper()

	items, ok := raw.([]any)
	if !ok {
		t.Fatalf("expected a sequence, got %T", raw)
	}

	out := make(map[string]bool, len(items))
	for _, item := range items {
		out[fmt.Sprint(item)] = true
	}

	return out
}

func enumValues(t *testing.T, raw any) []any {
	t.Helper()

	schema, ok := raw.(map[string]any)
	if !ok {
		t.Fatalf("expected a schema mapping, got %T", raw)
	}

	values, ok := schema["enum"].([]any)
	if !ok {
		t.Fatalf("schema has no enum list: %v", schema)
	}

	return values
}
