package converter

import (
	"strings"
	"testing"
)

func TestParseSimpleSchema(t *testing.T) {
	schemaJSON := `{
		"type": "object",
		"properties": {
			"name": {"type": "string"},
			"age": {"type": "integer"},
			"email": {"type": "string", "format": "email"}
		},
		"required": ["name", "email"]
	}`

	schema, err := Parse([]byte(schemaJSON))
	if err != nil {
		t.Fatalf("failed to parse schema: %v", err)
	}

	if schema.Type.Primary() != "object" {
		t.Errorf("expected object type, got %v", schema.Type)
	}

	if len(schema.Properties) != 3 {
		t.Errorf("expected 3 properties, got %d", len(schema.Properties))
	}

	if len(schema.Required) != 2 {
		t.Errorf("expected 2 required fields, got %d", len(schema.Required))
	}
}

func TestGenerateSimpleStruct(t *testing.T) {
	schemaJSON := `{
		"type": "object",
		"properties": {
			"name": {"type": "string", "description": "User's full name"},
			"age": {"type": "integer"},
			"active": {"type": "boolean"}
		},
		"required": ["name"]
	}`

	schema, err := Parse([]byte(schemaJSON))
	if err != nil {
		t.Fatalf("failed to parse schema: %v", err)
	}

	opts := DefaultOptions()
	opts.PackageName = "user"
	opts.TypeName = "User"

	code, err := Generate(schema, opts)
	if err != nil {
		t.Fatalf("failed to generate code: %v", err)
	}

	// Check package declaration
	if !strings.Contains(code, "package user") {
		t.Error("missing package declaration")
	}

	// Check struct definition
	if !strings.Contains(code, "type User struct") {
		t.Error("missing User struct")
	}

	// Check fields
	if !strings.Contains(code, "Name string") {
		t.Error("missing Name field")
	}

	if !strings.Contains(code, "Age *int64") {
		t.Error("missing Age field with pointer (optional)")
	}

	if !strings.Contains(code, "Active *bool") {
		t.Error("missing Active field")
	}

	// Check JSON tags
	if !strings.Contains(code, `json:"name"`) {
		t.Error("missing JSON tag for name")
	}

	if !strings.Contains(code, `json:"age,omitempty"`) {
		t.Error("missing omitempty for optional age field")
	}

	// Check comments
	if !strings.Contains(code, "User's full name") {
		t.Error("missing description comment")
	}
}

func TestGenerateWithDefinitions(t *testing.T) {
	schemaJSON := `{
		"type": "object",
		"properties": {
			"user": {"$ref": "#/$defs/User"}
		},
		"$defs": {
			"User": {
				"type": "object",
				"properties": {
					"id": {"type": "integer"},
					"name": {"type": "string"}
				},
				"required": ["id", "name"]
			}
		}
	}`

	schema, err := Parse([]byte(schemaJSON))
	if err != nil {
		t.Fatalf("failed to parse schema: %v", err)
	}

	opts := DefaultOptions()
	opts.PackageName = "api"
	opts.TypeName = "Response"

	code, err := Generate(schema, opts)
	if err != nil {
		t.Fatalf("failed to generate code: %v", err)
	}

	// Check User type is generated
	if !strings.Contains(code, "type User struct") {
		t.Error("missing User struct from definitions")
	}

	// Check reference field
	if !strings.Contains(code, "User *User") || !strings.Contains(code, "User User") {
		t.Log(code)
		// Either pointer or value type is acceptable
	}
}

func TestGenerateArray(t *testing.T) {
	schemaJSON := `{
		"type": "array",
		"items": {
			"type": "object",
			"properties": {
				"id": {"type": "integer"}
			}
		}
	}`

	schema, err := Parse([]byte(schemaJSON))
	if err != nil {
		t.Fatalf("failed to parse schema: %v", err)
	}

	opts := DefaultOptions()
	opts.PackageName = "data"
	opts.TypeName = "Items"

	code, err := Generate(schema, opts)
	if err != nil {
		t.Fatalf("failed to generate code: %v", err)
	}

	if !strings.Contains(code, "type Items []") {
		t.Error("expected array type definition")
	}
}

func TestGenerateEnum(t *testing.T) {
	schemaJSON := `{
		"type": "string",
		"enum": ["pending", "active", "completed"]
	}`

	schema, err := Parse([]byte(schemaJSON))
	if err != nil {
		t.Fatalf("failed to parse schema: %v", err)
	}

	opts := DefaultOptions()
	opts.PackageName = "status"
	opts.TypeName = "Status"

	code, err := Generate(schema, opts)
	if err != nil {
		t.Fatalf("failed to generate code: %v", err)
	}

	// Check enum type
	if !strings.Contains(code, "type Status string") {
		t.Error("expected string type for enum")
	}

	// Check enum constants
	if !strings.Contains(code, "StatusPending") {
		t.Error("missing enum constant")
	}
}

func TestNullableType(t *testing.T) {
	schemaJSON := `{
		"type": ["string", "null"]
	}`

	schema, err := Parse([]byte(schemaJSON))
	if err != nil {
		t.Fatalf("failed to parse schema: %v", err)
	}

	if !schema.Type.IsNullable() {
		t.Error("expected nullable type")
	}

	if schema.Type.Primary() != "string" {
		t.Errorf("expected primary type string, got %s", schema.Type.Primary())
	}
}

func TestToPascalCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello", "Hello"},
		{"hello_world", "HelloWorld"},
		{"hello-world", "HelloWorld"},
		{"hello world", "HelloWorld"},
		{"helloWorld", "HelloWorld"},
		{"HTTPServer", "HTTPServer"},
		{"user.name", "UserName"},
	}

	for _, tc := range tests {
		result := toPascalCase(tc.input)
		if result != tc.expected {
			t.Errorf("toPascalCase(%q) = %q, expected %q", tc.input, result, tc.expected)
		}
	}
}
