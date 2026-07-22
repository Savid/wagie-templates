package templatecheck

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// schemaVocabulary is the JSON-schema keyword set template schemas may use.
// YAML flow scalars end at commas, so an unquoted flow-mapping description
// containing a comma silently splits into a truncated description plus junk
// keys — any key outside this set is almost always a mangled description.
var schemaVocabulary = map[string]struct{}{
	"type": {}, "description": {}, "items": {}, "properties": {},
	"required": {}, "enum": {}, "default": {}, "const": {},
	"minimum": {}, "maximum": {}, "exclusiveMinimum": {}, "exclusiveMaximum": {},
	"multipleOf": {}, "minItems": {}, "maxItems": {}, "uniqueItems": {},
	"minLength": {}, "maxLength": {}, "pattern": {}, "format": {},
	"minProperties": {}, "maxProperties": {}, "additionalProperties": {},
	"title": {}, "examples": {}, "anyOf": {}, "oneOf": {}, "allOf": {},
	"not": {}, "$ref": {}, "$defs": {}, "definitions": {},
}

// StraySchemaKeyError reports one schema key outside the JSON-schema
// vocabulary — in practice the tail of an unquoted comma-containing
// description that YAML split into its own key.
type StraySchemaKeyError struct {
	Path string
	Line int
	Key  string
}

func (e StraySchemaKeyError) Error() string {
	return fmt.Sprintf(
		"%s:%d: schema key %q is not JSON-schema vocabulary (an unquoted description containing a comma splits at the comma in a flow mapping — quote the description)",
		e.Path, e.Line, e.Key,
	)
}

// StraySchemaKeys scans every `schema:` subtree in the given template files
// for keys outside the JSON-schema vocabulary, following aliases so anchored
// schemas are checked at their declaration. An anchored schema reached from
// several alias sites resolves to the same declaration nodes, so issues are
// deduplicated by site.
func StraySchemaKeys(files []SchemaSource) ([]StraySchemaKeyError, error) {
	issues := make([]StraySchemaKeyError, 0, 4)

	for _, file := range files {
		var root yaml.Node
		if err := yaml.Unmarshal(file.Data, &root); err != nil {
			return nil, fmt.Errorf("parse %s: %w", file.Path, err)
		}

		if len(root.Content) == 0 {
			continue
		}

		if err := walkForSchemaKeys(root.Content[0], file.Path, &issues, 0); err != nil {
			return nil, fmt.Errorf("scan %s: %w", file.Path, err)
		}
	}

	seen := make(map[StraySchemaKeyError]struct{}, len(issues))
	deduped := issues[:0]

	for _, issue := range issues {
		if _, ok := seen[issue]; ok {
			continue
		}

		seen[issue] = struct{}{}
		deduped = append(deduped, issue)
	}

	return deduped, nil
}

// walkForSchemaKeys descends the document and checks the value of every
// mapping entry keyed `schema` as a schema subtree.
func walkForSchemaKeys(node *yaml.Node, path string, issues *[]StraySchemaKeyError, depth int) error {
	if node == nil {
		return nil
	}

	if depth > maxShapeWalkDepth {
		return fmt.Errorf("schema key walk exceeded depth %d", maxShapeWalkDepth)
	}

	if node.Kind == yaml.AliasNode {
		return walkForSchemaKeys(node.Alias, path, issues, depth+1)
	}

	if node.Kind != yaml.MappingNode {
		for _, child := range node.Content {
			if err := walkForSchemaKeys(child, path, issues, depth+1); err != nil {
				return err
			}
		}

		return nil
	}

	for i := 0; i+1 < len(node.Content); i += 2 {
		key, value := node.Content[i], node.Content[i+1]

		checker := walkForSchemaKeys
		if key.Value == "schema" {
			checker = checkSchemaKeys
		}

		if err := checker(value, path, issues, depth+1); err != nil {
			return err
		}
	}

	return nil
}

// checkSchemaKeys validates one schema mapping's keys and recurses into the
// positions whose values are themselves schemas.
func checkSchemaKeys(node *yaml.Node, path string, issues *[]StraySchemaKeyError, depth int) error {
	if depth > maxShapeWalkDepth {
		return fmt.Errorf("schema key walk exceeded depth %d", maxShapeWalkDepth)
	}

	node = resolveAliasNode(node)
	if node == nil || node.Kind != yaml.MappingNode {
		return nil // e.g. `additionalProperties: false`
	}

	for i := 0; i+1 < len(node.Content); i += 2 {
		key, value := node.Content[i], node.Content[i+1]

		if _, ok := schemaVocabulary[key.Value]; !ok {
			*issues = append(*issues, StraySchemaKeyError{Path: path, Line: key.Line, Key: key.Value})

			continue
		}

		if err := checkSchemaChild(key.Value, value, path, issues, depth+1); err != nil {
			return err
		}
	}

	return nil
}

// checkSchemaChild recurses into the schema positions nested under one
// vocabulary keyword; keywords carrying plain values recurse nowhere.
func checkSchemaChild(keyword string, value *yaml.Node, path string, issues *[]StraySchemaKeyError, depth int) error {
	switch keyword {
	case "items", "additionalProperties", "not":
		return checkSchemaKeys(value, path, issues, depth)
	case "properties", "$defs", "definitions":
		return checkSchemaMapValues(value, path, issues, depth)
	case "anyOf", "oneOf", "allOf":
		return checkSchemaSeqItems(value, path, issues, depth)
	default:
		return nil
	}
}

// checkSchemaMapValues checks every value of a map-of-schemas node (e.g. the
// entries under `properties`).
func checkSchemaMapValues(node *yaml.Node, path string, issues *[]StraySchemaKeyError, depth int) error {
	node = resolveAliasNode(node)
	if node == nil || node.Kind != yaml.MappingNode {
		return nil
	}

	for i := 0; i+1 < len(node.Content); i += 2 {
		if err := checkSchemaKeys(node.Content[i+1], path, issues, depth); err != nil {
			return err
		}
	}

	return nil
}

// checkSchemaSeqItems checks every item of a list-of-schemas node (e.g. the
// entries under `anyOf`).
func checkSchemaSeqItems(node *yaml.Node, path string, issues *[]StraySchemaKeyError, depth int) error {
	node = resolveAliasNode(node)
	if node == nil || node.Kind != yaml.SequenceNode {
		return nil
	}

	for _, item := range node.Content {
		if err := checkSchemaKeys(item, path, issues, depth); err != nil {
			return err
		}
	}

	return nil
}
