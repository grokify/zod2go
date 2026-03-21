# Release Notes: zod2go v0.1.0

**Release Date:** 2026-03-21

## Overview

Initial release of zod2go, a CLI tool that converts TypeScript Zod schemas to Go struct definitions via JSON Schema intermediate representation.

## Highlights

- **Full Conversion Pipeline**: Convert Zod schemas directly to Go structs with a single command
- **JSON Schema Intermediate**: Use JSON Schema as an interop format for maximum flexibility
- **Insomnia v5 Example**: Complete working example converting Insomnia's export format schemas

## Features

### CLI Commands

| Command | Description |
|---------|-------------|
| `zod2go generate` | Full pipeline: Zod → JSON Schema → Go |
| `zod2go zod2json` | Convert Zod schema to JSON Schema |
| `zod2go json2go` | Convert JSON Schema to Go structs |
| `zod2go check` | Verify Node.js dependencies |

### JSON Schema Parser

- Parse JSON Schema from files or bytes
- Support for objects, arrays, enums, refs, nullable types
- Handle `$defs` and `definitions` sections
- Generate Go structs with `json` and optional `yaml` tags

### Go Code Generator

- PascalCase field names with proper capitalization
- Pointer types for optional fields (configurable)
- `omitempty` tags for optional JSON fields
- Comments from schema descriptions
- Enum constants generation

## Installation

```bash
# Install CLI
go install github.com/grokify/zod2go/cmd/zod2go@latest

# Verify installation
zod2go check
```

## Usage Example

```bash
# Full pipeline (Zod → Go)
zod2go generate -i schema.ts -o types_gen.go -p mypackage --export MySchema

# Step by step
zod2go zod2json -i schema.ts -o schema.json --export MySchema
zod2go json2go -i schema.json -o types_gen.go -p mypackage
```

## Requirements

- Go 1.21+
- Node.js 18+ with npx
- `zod-to-json-schema` npm package (or tsx/ts-node for TypeScript execution)

## Known Limitations

- Recursive JSON types degrade to `any` in Go
- Complex union types (discriminated unions) may need manual refinement
- Zod v3 required for zod-to-json-schema compatibility (v4 not yet supported)

## What's Next

- Improved union type handling
- Support for more JSON Schema features (allOf, oneOf)
- Go code formatting with gofmt/goimports
