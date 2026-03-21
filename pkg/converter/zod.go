package converter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// validateBinaryPath ensures the binary path ends with an expected binary name.
func validateBinaryPath(path string, expectedBinaries ...string) error {
	base := filepath.Base(path)
	for _, expected := range expectedBinaries {
		if base == expected {
			return nil
		}
	}
	return fmt.Errorf("invalid binary path %q: expected one of %v", path, expectedBinaries)
}

// ZodConvertOptions configures Zod to JSON Schema conversion.
type ZodConvertOptions struct {
	// ExportName is the name of the exported Zod schema to convert.
	ExportName string

	// RefStrategy controls how $refs are handled.
	// Options: "none", "root", "relative"
	RefStrategy string

	// WorkDir is the working directory for running Node.js.
	// Defaults to the directory containing the input file.
	WorkDir string

	// NodePath is the path to the Node.js executable.
	// Defaults to "node" (uses PATH).
	NodePath string

	// NpxPath is the path to npx executable.
	// Defaults to "npx" (uses PATH).
	NpxPath string
}

// DefaultZodConvertOptions returns default Zod conversion options.
func DefaultZodConvertOptions() ZodConvertOptions {
	return ZodConvertOptions{
		RefStrategy: "none",
		NodePath:    "node",
		NpxPath:     "npx",
	}
}

// ZodToJSONSchema converts a Zod schema file to JSON Schema.
func ZodToJSONSchema(inputPath string, opts ZodConvertOptions) (*JSONSchema, error) {
	data, err := ZodToJSONSchemaBytes(inputPath, opts)
	if err != nil {
		return nil, err
	}
	return Parse(data)
}

// ZodToJSONSchemaBytes converts a Zod schema file to JSON Schema bytes.
func ZodToJSONSchemaBytes(inputPath string, opts ZodConvertOptions) ([]byte, error) {
	absPath, err := filepath.Abs(inputPath)
	if err != nil {
		return nil, fmt.Errorf("resolving input path: %w", err)
	}

	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("input file not found: %s", absPath)
	}

	workDir := opts.WorkDir
	if workDir == "" {
		workDir = filepath.Dir(absPath)
	}

	exportName := opts.ExportName
	if exportName == "" {
		// Try to infer from filename
		base := filepath.Base(inputPath)
		exportName = strings.TrimSuffix(base, filepath.Ext(base))
		exportName = toPascalCase(exportName) + "Schema"
	}

	refStrategy := opts.RefStrategy
	if refStrategy == "" {
		refStrategy = "none"
	}

	// Create inline script
	script := fmt.Sprintf(`
const { zodToJsonSchema } = require('zod-to-json-schema');

async function main() {
  try {
    // Dynamic import for ESM/CJS compatibility
    const mod = await import(%q);
    const schema = mod.%s || mod.default?.%s || mod.default;

    if (!schema) {
      console.error('Schema not found: %s');
      console.error('Available exports:', Object.keys(mod));
      process.exit(1);
    }

    const jsonSchema = zodToJsonSchema(schema, {
      name: %q,
      $refStrategy: %q,
    });

    console.log(JSON.stringify(jsonSchema, null, 2));
  } catch (err) {
    console.error('Error:', err.message);
    process.exit(1);
  }
}

main();
`, absPath, exportName, exportName, exportName, exportName, refStrategy)

	// Validate binary paths before execution
	if err := validateBinaryPath(opts.NpxPath, "npx"); err != nil {
		return nil, fmt.Errorf("validating npx path: %w", err)
	}
	if err := validateBinaryPath(opts.NodePath, "node"); err != nil {
		return nil, fmt.Errorf("validating node path: %w", err)
	}

	// Try tsx first, then ts-node, then direct node
	var stdout, stderr bytes.Buffer

	// Try with npx tsx (handles TypeScript natively)
	cmd := exec.Command(opts.NpxPath, "tsx", "-e", script) // #nosec G204 - paths validated above
	cmd.Dir = workDir
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err == nil && stdout.Len() > 0 {
		return stdout.Bytes(), nil
	}

	// Reset buffers
	stdout.Reset()
	stderr.Reset()

	// Try with npx ts-node
	cmd = exec.Command(opts.NpxPath, "ts-node", "-e", script) // #nosec G204 - paths validated above
	cmd.Dir = workDir
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err == nil && stdout.Len() > 0 {
		return stdout.Bytes(), nil
	}

	// Reset buffers
	stdout.Reset()
	stderr.Reset()

	// Try with node directly (for JS files or if TypeScript is pre-compiled)
	cmd = exec.Command(opts.NodePath, "-e", script) // #nosec G204 - paths validated above
	cmd.Dir = workDir
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("failed to convert Zod schema: %v\nstderr: %s", err, stderr.String())
	}

	if stdout.Len() == 0 {
		return nil, fmt.Errorf("no output from Zod conversion\nstderr: %s", stderr.String())
	}

	// Validate it's valid JSON
	var js json.RawMessage
	if err := json.Unmarshal(stdout.Bytes(), &js); err != nil {
		return nil, fmt.Errorf("invalid JSON output: %w\noutput: %s", err, stdout.String())
	}

	return stdout.Bytes(), nil
}

// ZodToGo converts a Zod schema file directly to Go code.
func ZodToGo(inputPath string, zodOpts ZodConvertOptions, goOpts Options) (string, error) {
	schema, err := ZodToJSONSchema(inputPath, zodOpts)
	if err != nil {
		return "", fmt.Errorf("converting Zod to JSON Schema: %w", err)
	}

	return Generate(schema, goOpts)
}

// CheckNodeDependencies verifies that required Node.js dependencies are available.
func CheckNodeDependencies() error {
	// Check node
	if _, err := exec.LookPath("node"); err != nil {
		return fmt.Errorf("node not found in PATH: %w", err)
	}

	// Check npx
	if _, err := exec.LookPath("npx"); err != nil {
		return fmt.Errorf("npx not found in PATH: %w", err)
	}

	// Check zod-to-json-schema is available
	cmd := exec.Command("npx", "zod-to-json-schema", "--version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("zod-to-json-schema not available. Install with: npm install -g zod-to-json-schema")
	}

	return nil
}
