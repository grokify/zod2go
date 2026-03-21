package main

import (
	"fmt"
	"os"

	"github.com/grokify/zod2go/pkg/converter"
	"github.com/spf13/cobra"
)

var zod2jsonCmd = &cobra.Command{
	Use:   "zod2json",
	Short: "Convert Zod schema to JSON Schema",
	Long: `Convert a TypeScript Zod schema file to JSON Schema format.

This is step 1 of the conversion pipeline. The output JSON Schema can then
be converted to Go structs using the json2go command.

Requires Node.js and zod-to-json-schema to be installed.`,
	Example: `  # Convert to stdout
  zod2go zod2json -i schema.ts --export MySchema

  # Convert to file
  zod2go zod2json -i schema.ts -o schema.json --export MySchema`,
	RunE: runZod2Json,
}

var (
	z2jInput       string
	z2jOutput      string
	z2jExport      string
	z2jRefStrategy string
)

func init() {
	zod2jsonCmd.Flags().StringVarP(&z2jInput, "input", "i", "", "Input Zod schema file (.ts)")
	zod2jsonCmd.Flags().StringVarP(&z2jOutput, "output", "o", "", "Output JSON Schema file (default: stdout)")
	zod2jsonCmd.Flags().StringVarP(&z2jExport, "export", "e", "", "Exported Zod schema name to convert")
	zod2jsonCmd.Flags().StringVar(&z2jRefStrategy, "ref-strategy", "none", "Reference strategy: none, root, relative")

	if err := zod2jsonCmd.MarkFlagRequired("input"); err != nil {
		panic(fmt.Sprintf("failed to mark input flag required: %v", err))
	}
	if err := zod2jsonCmd.MarkFlagRequired("export"); err != nil {
		panic(fmt.Sprintf("failed to mark export flag required: %v", err))
	}
}

func runZod2Json(cmd *cobra.Command, args []string) error {
	opts := converter.DefaultZodConvertOptions()
	opts.ExportName = z2jExport
	opts.RefStrategy = z2jRefStrategy

	data, err := converter.ZodToJSONSchemaBytes(z2jInput, opts)
	if err != nil {
		return fmt.Errorf("conversion failed: %w", err)
	}

	if z2jOutput == "" || z2jOutput == "-" {
		fmt.Println(string(data))
	} else {
		if err := os.WriteFile(z2jOutput, data, 0600); err != nil {
			return fmt.Errorf("writing output file: %w", err)
		}
		fmt.Fprintf(os.Stderr, "Generated: %s\n", z2jOutput)
	}

	return nil
}
