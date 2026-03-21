# zod2go

Convert Zod schemas to Go structs via JSON Schema intermediate representation.

## Overview

`zod2go` provides a pipeline to convert TypeScript Zod schemas into Go struct definitions:

```
Zod Schema (.ts) → JSON Schema (.json) → Go Structs (.go)
```

This enables Go projects to consume type definitions from TypeScript/JavaScript projects that use Zod for validation.

## Installation

```bash
# Install the Go tool
go install github.com/grokify/zod2go/cmd/zod2go@latest

# Ensure Node.js dependencies are available (for Zod conversion)
npm install -g zod-to-json-schema
```

## Usage

### Full Pipeline (Zod → Go)

```bash
# Convert a Zod schema file to Go structs
zod2go generate -i schema.ts -o types_gen.go -p mypackage

# With custom export name
zod2go generate -i schema.ts -o types_gen.go -p mypackage --export MySchema
```

### Step-by-Step

```bash
# Step 1: Convert Zod to JSON Schema
zod2go zod2json -i schema.ts -o schema.json --export MySchema

# Step 2: Convert JSON Schema to Go
zod2go json2go -i schema.json -o types_gen.go -p mypackage
```

### Using the Node.js Script Directly

```bash
# For complex Zod schemas, use the extraction script
node scripts/zod-extract.mjs input.ts OutputSchema > schema.json
```

## Examples

### Insomnia Schema Extraction

```bash
# Clone Insomnia repo
git clone https://github.com/Kong/insomnia.git /tmp/insomnia

# Extract the v5 parser schema
zod2go generate \
  -i /tmp/insomnia/packages/insomnia/src/common/import-v5-parser.ts \
  -o pkg/insomnia/types_gen.go \
  -p insomnia \
  --export InsomniaFileSchema
```

## How It Works

1. **Zod → JSON Schema**: Uses `zod-to-json-schema` (Node.js) to convert Zod validation schemas to JSON Schema format
2. **JSON Schema → Go**: Uses `go-jsonschema` or built-in converter to generate Go struct definitions with JSON tags

## Supported Zod Types

| Zod Type | JSON Schema | Go Type |
|----------|-------------|---------|
| `z.string()` | `{"type": "string"}` | `string` |
| `z.number()` | `{"type": "number"}` | `float64` |
| `z.boolean()` | `{"type": "boolean"}` | `bool` |
| `z.array()` | `{"type": "array"}` | `[]T` |
| `z.object()` | `{"type": "object"}` | `struct` |
| `z.enum()` | `{"enum": [...]}` | `string` + constants |
| `z.union()` | `{"oneOf": [...]}` | `interface{}` or discriminated union |
| `z.optional()` | `{"required": false}` | `*T` (pointer) |
| `z.nullable()` | `{"nullable": true}` | `*T` (pointer) |

## Project Structure

```
zod2go/
├── cmd/zod2go/          # CLI application
├── pkg/converter/       # Go conversion logic
├── scripts/             # Node.js helper scripts
│   └── zod-extract.mjs  # Zod to JSON Schema extractor
├── examples/
│   └── insomnia/        # Insomnia schema example
└── README.md
```

## Requirements

- Go 1.21+
- Node.js 18+ (for Zod conversion)
- npm packages: `zod`, `zod-to-json-schema`, `typescript`, `ts-node`

## License

MIT
