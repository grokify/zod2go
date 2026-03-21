# Insomnia Schema Example

This example demonstrates converting Insomnia's Zod schemas to Go structs.

## Generated Files

| File | Description |
|------|-------------|
| `insomnia-schema.ts` | Standalone Zod schema extracted from Insomnia |
| `insomnia-v5.schema.json` | JSON Schema generated from Zod |
| `insomnia_gen.go` | Go structs generated from JSON Schema |

## Source

The schemas are extracted from Insomnia's source code:

- **Original**: `packages/insomnia/src/common/import-v5-parser.ts`
- **Repository**: https://github.com/Kong/insomnia

## Conversion Pipeline

```
Zod Schema (.ts) → JSON Schema (.json) → Go Structs (.go)
```

### Reproduce the Conversion

```bash
# Create isolated environment (Zod v3 required for zod-to-json-schema compatibility)
mkdir /tmp/zod-convert && cd /tmp/zod-convert
npm init -y
npm install zod@3 zod-to-json-schema typescript tsx

# Copy the standalone schema
cp examples/insomnia/insomnia-schema.ts .

# Create conversion script
cat > convert.ts << 'EOF'
import { zodToJsonSchema } from 'zod-to-json-schema';
import { CollectionSchema, ApiSpecSchema, ... } from './insomnia-schema';

// See the full script in the examples directory
EOF

# Run conversion
npx tsx convert.ts > insomnia-v5.schema.json

# Generate Go types
zod2go json2go -i insomnia-v5.schema.json -p insomnia -t InsomniaFile -o insomnia_gen.go
```

## Insomnia Export Formats

The Insomnia v5 format supports these file types (discriminated by `type` field):

| Type | Schema ID |
|------|-----------|
| Collection | `collection.insomnia.rest/5.0` |
| API Spec | `spec.insomnia.rest/5.0` |
| Mock Server | `mock.insomnia.rest/5.0` |
| Environment | `environment.insomnia.rest/5.0` |
| MCP Client | `mcpClient.insomnia/5.0` |

## Usage

```go
package main

import (
    "encoding/json"
    "os"
)

// InsomniaFile is a discriminated union - use type field to determine variant
type InsomniaFile = any

func main() {
    data, _ := os.ReadFile("export.json")

    // Parse as generic map first
    var raw map[string]any
    json.Unmarshal(data, &raw)

    // Check the type field
    switch raw["type"] {
    case "collection.insomnia.rest/5.0":
        var col Collection
        json.Unmarshal(data, &col)
        // handle collection
    case "spec.insomnia.rest/5.0":
        var spec ApiSpec
        json.Unmarshal(data, &spec)
        // handle API spec
    }
}
```

## Limitations

The auto-generated Go structs use `any` for several fields due to:

1. **Recursive JSON types**: The `JsonSchema` Zod type represents arbitrary JSON
2. **Complex union types**: Authentication schemas have 13+ variants with discriminator
3. **Nested object inlining**: Types are inlined rather than referenced

For production use, manually refine the types based on the JSON Schema.

## Related

- [Insomnia Import/Export Docs](https://docs.insomnia.rest/insomnia/import-export-data)
- [Insomnia v5 Parser Source](https://github.com/Kong/insomnia/blob/develop/packages/insomnia/src/common/import-v5-parser.ts)
