package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var version = "0.1.0"

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "zod2go",
	Short: "Convert Zod schemas to Go structs",
	Long: `zod2go converts TypeScript Zod schemas to Go struct definitions.

The conversion pipeline:
  Zod Schema (.ts) → JSON Schema (.json) → Go Structs (.go)

Examples:
  # Full pipeline: Zod to Go
  zod2go generate -i schema.ts -o types_gen.go -p mypackage --export MySchema

  # Step 1: Zod to JSON Schema
  zod2go zod2json -i schema.ts -o schema.json --export MySchema

  # Step 2: JSON Schema to Go
  zod2go json2go -i schema.json -o types_gen.go -p mypackage`,
	Version: version,
}

func init() {
	rootCmd.AddCommand(generateCmd)
	rootCmd.AddCommand(zod2jsonCmd)
	rootCmd.AddCommand(json2goCmd)
	rootCmd.AddCommand(checkCmd)
}
