package templatecheck

import (
	"encoding/json"
	"fmt"
	"slices"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// maxShapeWalkDepth bounds the alias-following walk; template schema trees are
// far shallower than this, so hitting it means a pathological document.
const maxShapeWalkDepth = 200

// sharedShape identifies a schema shape that templates copy between files
// because the spec grammar has no cross-file schema import. A schema node
// whose properties contain every marker key — and whose required list carries
// every requiredMarker, when set — is an occurrence of the shape.
type sharedShape struct {
	name            string
	markers         []string
	requiredMarkers []string
}

// sharedShapes are the runbook-owned contracts mirrored across templates (see
// ethereum/AGENTS.md). Marker sets must stay unique enough to not match
// unrelated objects. The network target shape additionally demands
// required:[kind] so the deliberately-relaxed caller-supplied-target bindings
// (provided_target, no required list) stay out of the comparison.
var sharedShapes = []sharedShape{
	{name: "issue record", markers: []string{"title", "summary", "fingerprint", "handles", "first_bad"}},
	{name: "evidence item", markers: []string{"source", "ref", "at", "detail"}},
	{name: "root-cause report", markers: []string{"root_cause", "trace_verdict", "review_verdict", "next_action"}},
	{name: "feedback queue", markers: []string{"priority_summary", "terminal", "terminal_reason", "tasks"}},
	{name: "network target", markers: []string{"kind", "enclave", "sandbox_id", "network_id"}, requiredMarkers: []string{"kind"}},
	{name: "snapshot catalog item", markers: []string{"epoch", "snapshot_id"}},
}

// SchemaSource is one template file to scan for shared schema occurrences.
type SchemaSource struct {
	Path string
	Data []byte
}

// SchemaDriftVariant is one distinct structure of a shared shape and every
// site declaring it.
type SchemaDriftVariant struct {
	Structure string
	Sites     []string
}

// SchemaDriftError reports a shared shape declared with more than one
// structure across the scanned files. Descriptions are ignored when comparing
// — vocabulary lives in descriptions and may legitimately vary per site.
type SchemaDriftError struct {
	Shape    string
	Variants []SchemaDriftVariant
}

func (e SchemaDriftError) Error() string {
	parts := make([]string, 0, len(e.Variants))
	for i, variant := range e.Variants {
		parts = append(parts, fmt.Sprintf("variant %d at %s", i+1, strings.Join(variant.Sites, ", ")))
	}
	return fmt.Sprintf(
		"shared schema %q has %d structural variants (descriptions ignored): %s",
		e.Shape,
		len(e.Variants),
		strings.Join(parts, "; "),
	)
}

// SchemaDrift compares every occurrence of each shared shape across the given
// template files and reports shapes that drifted into multiple structures.
func SchemaDrift(files []SchemaSource) ([]SchemaDriftError, error) {
	// shape name -> canonical structure -> site set
	occurrences := make(map[string]map[string]map[string]struct{}, len(sharedShapes))

	for _, file := range files {
		var root yaml.Node
		if err := yaml.Unmarshal(file.Data, &root); err != nil {
			return nil, fmt.Errorf("parse %s: %w", file.Path, err)
		}
		if len(root.Content) == 0 {
			continue
		}
		if err := walkForShapes(root.Content[0], file.Path, occurrences, 0); err != nil {
			return nil, fmt.Errorf("scan %s: %w", file.Path, err)
		}
	}

	drifted := make([]SchemaDriftError, 0, len(sharedShapes))
	for _, shape := range sharedShapes {
		structures := occurrences[shape.name]
		if len(structures) < 2 {
			continue
		}
		variants := make([]SchemaDriftVariant, 0, len(structures))
		for structure, sites := range structures {
			sorted := make([]string, 0, len(sites))
			for site := range sites {
				sorted = append(sorted, site)
			}
			sort.Strings(sorted)
			variants = append(variants, SchemaDriftVariant{Structure: structure, Sites: sorted})
		}
		sort.Slice(variants, func(i, j int) bool { return variants[i].Sites[0] < variants[j].Sites[0] })
		drifted = append(drifted, SchemaDriftError{Shape: shape.name, Variants: variants})
	}
	return drifted, nil
}

// walkForShapes records every mapping node matching a shared shape, following
// aliases so anchored schemas are seen at each use site.
func walkForShapes(node *yaml.Node, path string, occurrences map[string]map[string]map[string]struct{}, depth int) error {
	if node == nil {
		return nil
	}
	if depth > maxShapeWalkDepth {
		return fmt.Errorf("schema walk exceeded depth %d", maxShapeWalkDepth)
	}
	if node.Kind == yaml.AliasNode {
		return walkForShapes(node.Alias, path, occurrences, depth+1)
	}

	if node.Kind == yaml.MappingNode {
		if err := recordShapeOccurrences(node, path, occurrences); err != nil {
			return err
		}
	}

	for _, child := range node.Content {
		if err := walkForShapes(child, path, occurrences, depth+1); err != nil {
			return err
		}
	}
	return nil
}

// recordShapeOccurrences matches one mapping node against the shared shapes
// and records its canonical structure per matched shape.
func recordShapeOccurrences(node *yaml.Node, path string, occurrences map[string]map[string]map[string]struct{}) error {
	properties := resolveAliasNode(findMappingValue(node, "properties"))
	if properties == nil || properties.Kind != yaml.MappingNode {
		return nil
	}

	keys := mappingKeys(properties)
	required := sequenceValues(resolveAliasNode(findMappingValue(node, "required")))
	for _, shape := range sharedShapes {
		if !containsAll(keys, shape.markers) || !containsAll(required, shape.requiredMarkers) {
			continue
		}
		structure, err := canonicalStructure(node)
		if err != nil {
			return fmt.Errorf("normalize %s occurrence at line %d: %w", shape.name, node.Line, err)
		}
		if occurrences[shape.name] == nil {
			occurrences[shape.name] = make(map[string]map[string]struct{}, 2)
		}
		if occurrences[shape.name][structure] == nil {
			occurrences[shape.name][structure] = make(map[string]struct{}, 4)
		}
		occurrences[shape.name][structure][fmt.Sprintf("%s:%d", path, node.Line)] = struct{}{}
	}
	return nil
}

// canonicalStructure normalizes a schema node to canonical JSON with every
// description field removed, so only structural differences compare unequal.
func canonicalStructure(node *yaml.Node) (string, error) {
	var value any
	if err := node.Decode(&value); err != nil {
		return "", err
	}
	encoded, err := json.Marshal(stripDescriptions(value))
	if err != nil {
		return "", err
	}
	return string(encoded), nil
}

func stripDescriptions(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		out := make(map[string]any, len(typed))
		for key, entry := range typed {
			if key == "description" {
				continue
			}
			out[key] = stripDescriptions(entry)
		}
		return out
	case []any:
		out := make([]any, 0, len(typed))
		for _, entry := range typed {
			out = append(out, stripDescriptions(entry))
		}
		return out
	default:
		return value
	}
}

func resolveAliasNode(node *yaml.Node) *yaml.Node {
	if node != nil && node.Kind == yaml.AliasNode {
		return node.Alias
	}
	return node
}

func mappingKeys(node *yaml.Node) []string {
	keys := make([]string, 0, len(node.Content)/2)
	for i := 0; i+1 < len(node.Content); i += 2 {
		if node.Content[i].Kind == yaml.ScalarNode {
			keys = append(keys, node.Content[i].Value)
		}
	}
	return keys
}

func sequenceValues(node *yaml.Node) []string {
	if node == nil || node.Kind != yaml.SequenceNode {
		return nil
	}
	values := make([]string, 0, len(node.Content))
	for _, item := range node.Content {
		if item.Kind == yaml.ScalarNode {
			values = append(values, item.Value)
		}
	}
	return values
}

func containsAll(keys, markers []string) bool {
	for _, marker := range markers {
		if !slices.Contains(keys, marker) {
			return false
		}
	}
	return true
}
