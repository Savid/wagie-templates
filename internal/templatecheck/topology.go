package templatecheck

import (
	"fmt"
	"slices"
	"strings"

	"gopkg.in/yaml.v3"
)

// ConditionalNeedsError reports a conditional task that is used as a hard
// dependency by a sibling task. In Wagie, an if:false task settles skipped, and
// skipped needs cascade to dependents.
type ConditionalNeedsError struct {
	Path       string
	Task       string
	Dependents []string
	Line       int
}

func (i ConditionalNeedsError) Error() string {
	return fmt.Sprintf(
		"%s.%s has if: and is needed by %s; make it always complete with no-op outputs or remove it from downstream needs",
		i.Path,
		i.Task,
		strings.Join(i.Dependents, ", "),
	)
}

// ConditionalNeeds finds if-gated tasks that have dependent siblings in any
// tasks: block in a template.
func ConditionalNeeds(data []byte) ([]ConditionalNeedsError, error) {
	var root yaml.Node
	if err := yaml.Unmarshal(data, &root); err != nil {
		return nil, err
	}
	if len(root.Content) == 0 {
		return nil, nil
	}

	var issues []ConditionalNeedsError
	walkForTasks(root.Content[0], "template", &issues)
	return issues, nil
}

func walkForTasks(node *yaml.Node, path string, issues *[]ConditionalNeedsError) {
	if node == nil {
		return
	}
	if node.Kind != yaml.MappingNode {
		for _, child := range node.Content {
			walkForTasks(child, path, issues)
		}
		return
	}

	for i := 0; i+1 < len(node.Content); i += 2 {
		key, value := node.Content[i], node.Content[i+1]
		if key.Value == "tasks" && value.Kind == yaml.MappingNode {
			checkTaskSet(value, path+".tasks", issues)
		}
		walkForTasks(value, path+"."+key.Value, issues)
	}
}

func checkTaskSet(tasks *yaml.Node, path string, issues *[]ConditionalNeedsError) {
	conditional := map[string]int{}
	dependents := map[string][]string{}

	for i := 0; i+1 < len(tasks.Content); i += 2 {
		nameNode, taskNode := tasks.Content[i], tasks.Content[i+1]
		name := nameNode.Value
		if findMappingValue(taskNode, "if") != nil {
			conditional[name] = nameNode.Line
		}
		for _, need := range needsOf(taskNode) {
			dependents[need] = append(dependents[need], name)
		}
	}

	for task, line := range conditional {
		users := dependents[task]
		if len(users) == 0 {
			continue
		}
		slices.Sort(users)
		*issues = append(*issues, ConditionalNeedsError{
			Path:       path,
			Task:       task,
			Dependents: users,
			Line:       line,
		})
	}
}

func needsOf(task *yaml.Node) []string {
	needs := findMappingValue(task, "needs")
	if needs == nil {
		return nil
	}
	switch needs.Kind {
	case yaml.SequenceNode:
		out := make([]string, 0, len(needs.Content))
		for _, item := range needs.Content {
			if item.Kind == yaml.ScalarNode && item.Value != "" {
				out = append(out, item.Value)
			}
		}
		return out
	case yaml.ScalarNode:
		if needs.Value != "" {
			return []string{needs.Value}
		}
	case yaml.DocumentNode, yaml.MappingNode, yaml.AliasNode:
		return nil
	}
	return nil
}

func findMappingValue(node *yaml.Node, key string) *yaml.Node {
	if node == nil || node.Kind != yaml.MappingNode {
		return nil
	}
	for i := 0; i+1 < len(node.Content); i += 2 {
		if node.Content[i].Value == key {
			return node.Content[i+1]
		}
	}
	return nil
}
