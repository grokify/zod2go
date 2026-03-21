// Package converter provides JSON Schema to Go struct conversion.
package converter

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"unicode"
)

// JSONSchema represents a JSON Schema document.
type JSONSchema struct {
	Schema      string                 `json:"$schema,omitempty"`
	ID          string                 `json:"$id,omitempty"`
	Ref         string                 `json:"$ref,omitempty"`
	Title       string                 `json:"title,omitempty"`
	Description string                 `json:"description,omitempty"`
	Type        JSONType               `json:"type,omitempty"`
	Properties  map[string]*JSONSchema `json:"properties,omitempty"`
	Items       *JSONSchema            `json:"items,omitempty"`
	Required    []string               `json:"required,omitempty"`
	Enum        []any                  `json:"enum,omitempty"`
	Const       any                    `json:"const,omitempty"`
	OneOf       []*JSONSchema          `json:"oneOf,omitempty"`
	AnyOf       []*JSONSchema          `json:"anyOf,omitempty"`
	AllOf       []*JSONSchema          `json:"allOf,omitempty"`
	Definitions map[string]*JSONSchema `json:"definitions,omitempty"`
	Defs        map[string]*JSONSchema `json:"$defs,omitempty"`

	// String validation
	MinLength *int   `json:"minLength,omitempty"`
	MaxLength *int   `json:"maxLength,omitempty"`
	Pattern   string `json:"pattern,omitempty"`
	Format    string `json:"format,omitempty"`

	// Number validation
	Minimum          *float64 `json:"minimum,omitempty"`
	Maximum          *float64 `json:"maximum,omitempty"`
	ExclusiveMinimum *float64 `json:"exclusiveMinimum,omitempty"`
	ExclusiveMaximum *float64 `json:"exclusiveMaximum,omitempty"`
	MultipleOf       *float64 `json:"multipleOf,omitempty"`

	// Array validation
	MinItems    *int `json:"minItems,omitempty"`
	MaxItems    *int `json:"maxItems,omitempty"`
	UniqueItems bool `json:"uniqueItems,omitempty"`

	// Object validation
	AdditionalProperties any  `json:"additionalProperties,omitempty"`
	MinProperties        *int `json:"minProperties,omitempty"`
	MaxProperties        *int `json:"maxProperties,omitempty"`

	// Nullable
	Nullable bool `json:"nullable,omitempty"`

	// Default
	Default any `json:"default,omitempty"`
}

// JSONType handles JSON Schema type which can be string or array of strings.
type JSONType []string

// UnmarshalJSON implements custom unmarshaling for type field.
func (t *JSONType) UnmarshalJSON(data []byte) error {
	// Try string first
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		*t = []string{s}
		return nil
	}

	// Try array
	var arr []string
	if err := json.Unmarshal(data, &arr); err == nil {
		*t = arr
		return nil
	}

	return fmt.Errorf("type must be string or array of strings")
}

// HasType checks if the schema has a specific type.
func (t JSONType) HasType(typ string) bool {
	for _, v := range t {
		if v == typ {
			return true
		}
	}
	return false
}

// Primary returns the primary (non-null) type.
func (t JSONType) Primary() string {
	for _, v := range t {
		if v != "null" {
			return v
		}
	}
	if len(t) > 0 {
		return t[0]
	}
	return ""
}

// IsNullable returns true if the type includes "null".
func (t JSONType) IsNullable() bool {
	return t.HasType("null")
}

// Options configures the Go code generator.
type Options struct {
	PackageName    string
	TypeName       string // Root type name
	GenerateJSON   bool   // Generate json tags
	GenerateYAML   bool   // Generate yaml tags
	OmitEmpty      bool   // Add omitempty to optional fields
	UsePointers    bool   // Use pointers for optional fields
	InlineAllOf    bool   // Inline allOf schemas instead of embedding
	GenerateEnums  bool   // Generate const blocks for enums
	AddComments    bool   // Add comments from descriptions
	ExportedFields bool   // Export all struct fields (capitalize)
}

// DefaultOptions returns default generator options.
func DefaultOptions() Options {
	return Options{
		PackageName:    "types",
		TypeName:       "Root",
		GenerateJSON:   true,
		GenerateYAML:   false,
		OmitEmpty:      true,
		UsePointers:    true,
		InlineAllOf:    true,
		GenerateEnums:  true,
		AddComments:    true,
		ExportedFields: true,
	}
}

// TypeNameFromExport derives a Go type name from a Zod export name.
// It removes the "Schema" suffix if present.
func TypeNameFromExport(exportName string) string {
	if exportName == "" {
		return "Root"
	}
	name := exportName
	const suffix = "Schema"
	if len(name) > len(suffix) && name[len(name)-len(suffix):] == suffix {
		name = name[:len(name)-len(suffix)]
	}
	return name
}

// Generator converts JSON Schema to Go code.
type Generator struct {
	opts        Options
	schema      *JSONSchema
	definitions map[string]*JSONSchema
	generated   map[string]bool
	output      strings.Builder
	types       []string // Generated type definitions
	enums       []string // Generated enum constants
}

// NewGenerator creates a new generator.
func NewGenerator(schema *JSONSchema, opts Options) *Generator {
	g := &Generator{
		opts:        opts,
		schema:      schema,
		definitions: make(map[string]*JSONSchema),
		generated:   make(map[string]bool),
	}

	// Collect definitions from both locations
	for name, def := range schema.Definitions {
		g.definitions[name] = def
	}
	for name, def := range schema.Defs {
		g.definitions[name] = def
	}

	return g
}

// Generate produces Go code from the schema.
func (g *Generator) Generate() (string, error) {
	g.output.Reset()
	g.types = nil
	g.enums = nil

	// Generate header
	g.output.WriteString(fmt.Sprintf("// Code generated by zod2go. DO NOT EDIT.\n\n"))
	g.output.WriteString(fmt.Sprintf("package %s\n\n", g.opts.PackageName))

	// Generate root type
	g.generateType(g.opts.TypeName, g.schema)

	// Generate definitions
	names := make([]string, 0, len(g.definitions))
	for name := range g.definitions {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		if g.generated[name] {
			continue
		}
		g.generateType(name, g.definitions[name])
	}

	// Write types
	for _, t := range g.types {
		g.output.WriteString(t)
		g.output.WriteString("\n")
	}

	// Write enums
	if len(g.enums) > 0 {
		g.output.WriteString("// Enum values\nconst (\n")
		for _, e := range g.enums {
			g.output.WriteString(e)
		}
		g.output.WriteString(")\n")
	}

	return g.output.String(), nil
}

// generateType generates a Go type from a JSON Schema.
func (g *Generator) generateType(name string, schema *JSONSchema) {
	if g.generated[name] {
		return
	}
	g.generated[name] = true

	typeName := g.toTypeName(name)

	var buf strings.Builder

	// Add comment
	if g.opts.AddComments && schema.Description != "" {
		buf.WriteString(fmt.Sprintf("// %s %s\n", typeName, cleanComment(schema.Description)))
	}

	primaryType := schema.Type.Primary()

	switch {
	case schema.Ref != "":
		// Reference to another type
		refName := extractRefName(schema.Ref)
		buf.WriteString(fmt.Sprintf("type %s = %s\n", typeName, g.toTypeName(refName)))

	case len(schema.Enum) > 0:
		// Enum type
		goType := g.schemaToGoType(schema, false)
		buf.WriteString(fmt.Sprintf("type %s %s\n", typeName, goType))

		if g.opts.GenerateEnums {
			g.generateEnumConsts(typeName, schema.Enum)
		}

	case len(schema.OneOf) > 0 || len(schema.AnyOf) > 0:
		// Union type - generate interface or discriminated union
		buf.WriteString(fmt.Sprintf("type %s any // Union type - see oneOf/anyOf variants\n", typeName))

	case primaryType == "object" || len(schema.Properties) > 0:
		// Object type - generate struct
		buf.WriteString(fmt.Sprintf("type %s struct {\n", typeName))

		requiredSet := make(map[string]bool)
		for _, r := range schema.Required {
			requiredSet[r] = true
		}

		// Sort property names for deterministic output
		propNames := make([]string, 0, len(schema.Properties))
		for name := range schema.Properties {
			propNames = append(propNames, name)
		}
		sort.Strings(propNames)

		for _, propName := range propNames {
			propSchema := schema.Properties[propName]
			fieldName := g.toFieldName(propName)
			goType := g.schemaToGoType(propSchema, !requiredSet[propName])

			// Build tags
			var tags []string
			if g.opts.GenerateJSON {
				jsonTag := propName
				if g.opts.OmitEmpty && !requiredSet[propName] {
					jsonTag += ",omitempty"
				}
				tags = append(tags, fmt.Sprintf(`json:"%s"`, jsonTag))
			}
			if g.opts.GenerateYAML {
				yamlTag := propName
				if g.opts.OmitEmpty && !requiredSet[propName] {
					yamlTag += ",omitempty"
				}
				tags = append(tags, fmt.Sprintf(`yaml:"%s"`, yamlTag))
			}

			tagStr := ""
			if len(tags) > 0 {
				tagStr = fmt.Sprintf(" `%s`", strings.Join(tags, " "))
			}

			// Add field comment
			comment := ""
			if g.opts.AddComments && propSchema.Description != "" {
				comment = " // " + cleanComment(propSchema.Description)
			}

			buf.WriteString(fmt.Sprintf("\t%s %s%s%s\n", fieldName, goType, tagStr, comment))
		}

		buf.WriteString("}\n")

	case primaryType == "array":
		// Array type
		itemType := "any"
		if schema.Items != nil {
			itemType = g.schemaToGoType(schema.Items, false)
		}
		buf.WriteString(fmt.Sprintf("type %s []%s\n", typeName, itemType))

	default:
		// Simple type alias
		goType := g.schemaToGoType(schema, false)
		buf.WriteString(fmt.Sprintf("type %s %s\n", typeName, goType))
	}

	g.types = append(g.types, buf.String())
}

// schemaToGoType converts a JSON Schema to a Go type string.
func (g *Generator) schemaToGoType(schema *JSONSchema, optional bool) string {
	if schema == nil {
		return "any"
	}

	// Handle references
	if schema.Ref != "" {
		refName := extractRefName(schema.Ref)
		typeName := g.toTypeName(refName)
		if optional && g.opts.UsePointers {
			return "*" + typeName
		}
		return typeName
	}

	// Handle nullable
	nullable := schema.Nullable || schema.Type.IsNullable()

	primaryType := schema.Type.Primary()

	var goType string

	switch primaryType {
	case "string":
		goType = "string"
	case "integer":
		goType = "int64"
	case "number":
		goType = "float64"
	case "boolean":
		goType = "bool"
	case "array":
		itemType := "any"
		if schema.Items != nil {
			itemType = g.schemaToGoType(schema.Items, false)
		}
		goType = "[]" + itemType
	case "object":
		if len(schema.Properties) == 0 {
			// Map type
			goType = "map[string]any"
		} else {
			// Anonymous struct or reference
			goType = "any" // Could generate inline struct
		}
	default:
		goType = "any"
	}

	// Apply pointer for optional/nullable
	if (optional || nullable) && g.opts.UsePointers && !strings.HasPrefix(goType, "[]") && !strings.HasPrefix(goType, "map") {
		goType = "*" + goType
	}

	return goType
}

// generateEnumConsts generates const block for enum values.
func (g *Generator) generateEnumConsts(typeName string, values []any) {
	for _, v := range values {
		constName := fmt.Sprintf("%s%s", typeName, g.toTypeName(fmt.Sprintf("%v", v)))
		constValue := fmt.Sprintf("%q", v)
		g.enums = append(g.enums, fmt.Sprintf("\t%s %s = %s\n", constName, typeName, constValue))
	}
}

// toTypeName converts a string to a Go type name.
func (g *Generator) toTypeName(s string) string {
	return toPascalCase(s)
}

// toFieldName converts a string to a Go field name.
func (g *Generator) toFieldName(s string) string {
	name := toPascalCase(s)
	if g.opts.ExportedFields {
		return name
	}
	// Make first letter lowercase
	if len(name) > 0 {
		return strings.ToLower(name[:1]) + name[1:]
	}
	return name
}

// toPascalCase converts a string to PascalCase.
func toPascalCase(s string) string {
	var result strings.Builder
	capitalizeNext := true

	for _, r := range s {
		if r == '_' || r == '-' || r == ' ' || r == '.' {
			capitalizeNext = true
			continue
		}

		if capitalizeNext {
			result.WriteRune(unicode.ToUpper(r))
			capitalizeNext = false
		} else {
			result.WriteRune(r)
		}
	}

	return result.String()
}

// extractRefName extracts the type name from a $ref.
func extractRefName(ref string) string {
	// Handle "#/$defs/Name" or "#/definitions/Name"
	parts := strings.Split(ref, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return ref
}

// cleanComment cleans a description for use as a Go comment.
func cleanComment(s string) string {
	// Replace newlines with spaces
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", "")
	// Trim and collapse whitespace
	s = strings.Join(strings.Fields(s), " ")
	return s
}

// ParseFile reads and parses a JSON Schema file.
func ParseFile(path string) (*JSONSchema, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return Parse(data)
}

// Parse parses JSON Schema from bytes.
func Parse(data []byte) (*JSONSchema, error) {
	var schema JSONSchema
	if err := json.Unmarshal(data, &schema); err != nil {
		return nil, err
	}
	return &schema, nil
}

// GenerateFromFile generates Go code from a JSON Schema file.
func GenerateFromFile(path string, opts Options) (string, error) {
	schema, err := ParseFile(path)
	if err != nil {
		return "", err
	}
	return Generate(schema, opts)
}

// Generate generates Go code from a JSON Schema.
func Generate(schema *JSONSchema, opts Options) (string, error) {
	gen := NewGenerator(schema, opts)
	return gen.Generate()
}
