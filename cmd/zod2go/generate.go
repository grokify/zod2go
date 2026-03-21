package main

import (
	"fmt"
	"os"

	"github.com/grokify/zod2go/pkg/converter"
	"github.com/spf13/cobra"
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Convert Zod schema to Go structs (full pipeline)",
	Long: `Convert a TypeScript Zod schema file directly to Go struct definitions.

This command runs the full pipeline:
  1. Zod Schema → JSON Schema (via zod-to-json-schema)
  2. JSON Schema → Go Structs

Requires Node.js and zod-to-json-schema to be installed.`,
	Example: `  # Basic usage
  zod2go generate -i schema.ts -o types_gen.go -p mypackage --export MySchema

  # From Insomnia parser
  zod2go generate -i import-v5-parser.ts -o insomnia.go -p insomnia --export InsomniaFileSchema`,
	RunE: runGenerate,
}

var (
	genInput      string
	genOutput     string
	genPackage    string
	genExport     string
	genTypeName   string
	genYAML       bool
	genNoPointers bool
	genSchemaOut  string
)

func init() {
	generateCmd.Flags().StringVarP(&genInput, "input", "i", "", "Input Zod schema file (.ts)")
	generateCmd.Flags().StringVarP(&genOutput, "output", "o", "", "Output Go file (default: stdout)")
	generateCmd.Flags().StringVarP(&genPackage, "package", "p", "types", "Go package name")
	generateCmd.Flags().StringVarP(&genExport, "export", "e", "", "Exported Zod schema name to convert")
	generateCmd.Flags().StringVarP(&genTypeName, "type", "t", "", "Root Go type name (default: from export name)")
	generateCmd.Flags().BoolVar(&genYAML, "yaml", false, "Generate YAML tags in addition to JSON")
	generateCmd.Flags().BoolVar(&genNoPointers, "no-pointers", false, "Don't use pointers for optional fields")
	generateCmd.Flags().StringVar(&genSchemaOut, "schema-out", "", "Also save intermediate JSON Schema to this file")

	if err := generateCmd.MarkFlagRequired("input"); err != nil {
		panic(fmt.Sprintf("failed to mark input flag required: %v", err))
	}
	if err := generateCmd.MarkFlagRequired("export"); err != nil {
		panic(fmt.Sprintf("failed to mark export flag required: %v", err))
	}
}

func runGenerate(cmd *cobra.Command, args []string) error {
	// Configure Zod conversion
	zodOpts := converter.DefaultZodConvertOptions()
	zodOpts.ExportName = genExport

	// Configure Go generation
	goOpts := converter.DefaultOptions()
	goOpts.PackageName = genPackage
	goOpts.GenerateYAML = genYAML
	goOpts.UsePointers = !genNoPointers

	if genTypeName != "" {
		goOpts.TypeName = genTypeName
	} else if genExport != "" {
		goOpts.TypeName = converter.TypeNameFromExport(genExport)
	}

	// Step 1: Convert Zod to JSON Schema
	jsonSchemaBytes, err := converter.ZodToJSONSchemaBytes(genInput, zodOpts)
	if err != nil {
		return fmt.Errorf("Zod to JSON Schema conversion failed: %w", err)
	}

	// Save JSON Schema if requested
	if genSchemaOut != "" {
		if err := os.WriteFile(genSchemaOut, jsonSchemaBytes, 0600); err != nil {
			return fmt.Errorf("writing JSON Schema file: %w", err)
		}
		fmt.Fprintf(os.Stderr, "JSON Schema: %s\n", genSchemaOut)
	}

	// Step 2: Parse JSON Schema
	schema, err := converter.Parse(jsonSchemaBytes)
	if err != nil {
		return fmt.Errorf("parsing JSON Schema: %w", err)
	}

	// Step 3: Generate Go code
	code, err := converter.Generate(schema, goOpts)
	if err != nil {
		return fmt.Errorf("Go generation failed: %w", err)
	}

	// Write output
	if genOutput == "" || genOutput == "-" {
		fmt.Print(code)
	} else {
		if err := os.WriteFile(genOutput, []byte(code), 0600); err != nil {
			return fmt.Errorf("writing output file: %w", err)
		}
		fmt.Fprintf(os.Stderr, "Generated: %s\n", genOutput)
	}

	return nil
}
