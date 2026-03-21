package main

import (
	"fmt"
	"os"

	"github.com/grokify/zod2go/pkg/converter"
	"github.com/spf13/cobra"
)

var json2goCmd = &cobra.Command{
	Use:   "json2go",
	Short: "Convert JSON Schema to Go structs",
	Long: `Convert a JSON Schema file to Go struct definitions.

This is step 2 of the conversion pipeline. Use this command after
converting Zod schemas to JSON Schema with the zod2json command.`,
	Example: `  # Convert to stdout
  zod2go json2go -i schema.json -p mypackage

  # Convert to file
  zod2go json2go -i schema.json -o types_gen.go -p mypackage

  # With custom type name
  zod2go json2go -i schema.json -o types.go -p mypackage -t Config`,
	RunE: runJson2Go,
}

var (
	j2gInput      string
	j2gOutput     string
	j2gPackage    string
	j2gTypeName   string
	j2gYAML       bool
	j2gNoPointers bool
	j2gNoComments bool
)

func init() {
	json2goCmd.Flags().StringVarP(&j2gInput, "input", "i", "", "Input JSON Schema file")
	json2goCmd.Flags().StringVarP(&j2gOutput, "output", "o", "", "Output Go file (default: stdout)")
	json2goCmd.Flags().StringVarP(&j2gPackage, "package", "p", "types", "Go package name")
	json2goCmd.Flags().StringVarP(&j2gTypeName, "type", "t", "Root", "Root Go type name")
	json2goCmd.Flags().BoolVar(&j2gYAML, "yaml", false, "Generate YAML tags in addition to JSON")
	json2goCmd.Flags().BoolVar(&j2gNoPointers, "no-pointers", false, "Don't use pointers for optional fields")
	json2goCmd.Flags().BoolVar(&j2gNoComments, "no-comments", false, "Don't generate comments from descriptions")

	if err := json2goCmd.MarkFlagRequired("input"); err != nil {
		panic(fmt.Sprintf("failed to mark input flag required: %v", err))
	}
}

func runJson2Go(cmd *cobra.Command, args []string) error {
	opts := converter.DefaultOptions()
	opts.PackageName = j2gPackage
	opts.TypeName = j2gTypeName
	opts.GenerateYAML = j2gYAML
	opts.UsePointers = !j2gNoPointers
	opts.AddComments = !j2gNoComments

	code, err := converter.GenerateFromFile(j2gInput, opts)
	if err != nil {
		return fmt.Errorf("conversion failed: %w", err)
	}

	if j2gOutput == "" || j2gOutput == "-" {
		fmt.Print(code)
	} else {
		if err := os.WriteFile(j2gOutput, []byte(code), 0600); err != nil {
			return fmt.Errorf("writing output file: %w", err)
		}
		fmt.Fprintf(os.Stderr, "Generated: %s\n", j2gOutput)
	}

	return nil
}
