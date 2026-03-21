package main

import (
	"fmt"
	"os/exec"

	"github.com/spf13/cobra"
)

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Check Node.js dependencies are installed",
	Long: `Verify that all required Node.js dependencies are available for Zod conversion.

Required dependencies:
  - node (Node.js runtime)
  - npx (Node package executor)
  - tsx or ts-node (TypeScript execution)
  - zod-to-json-schema (conversion library)`,
	RunE: runCheck,
}

func runCheck(cmd *cobra.Command, args []string) error {
	fmt.Println("Checking dependencies...")
	fmt.Println()

	allOk := true

	// Check node
	if path, err := exec.LookPath("node"); err != nil {
		fmt.Println("  node: NOT FOUND")
		allOk = false
	} else {
		version, _ := exec.Command("node", "--version").Output()
		fmt.Printf("  node: %s (%s)\n", string(version[:len(version)-1]), path)
	}

	// Check npx
	if path, err := exec.LookPath("npx"); err != nil {
		fmt.Println("  npx: NOT FOUND")
		allOk = false
	} else {
		fmt.Printf("  npx: found (%s)\n", path)
	}

	// Check tsx
	if err := exec.Command("npx", "tsx", "--version").Run(); err != nil {
		fmt.Println("  tsx: NOT FOUND (optional, will try ts-node)")
	} else {
		fmt.Println("  tsx: available")
	}

	// Check ts-node
	if err := exec.Command("npx", "ts-node", "--version").Run(); err != nil {
		fmt.Println("  ts-node: NOT FOUND (optional if tsx is available)")
	} else {
		fmt.Println("  ts-node: available")
	}

	// Check zod-to-json-schema
	checkScript := `try { require('zod-to-json-schema'); console.log('ok'); } catch(e) { process.exit(1); }`
	if err := exec.Command("node", "-e", checkScript).Run(); err != nil {
		fmt.Println("  zod-to-json-schema: NOT FOUND")
		fmt.Println("    Install with: npm install -g zod-to-json-schema")
		allOk = false
	} else {
		fmt.Println("  zod-to-json-schema: available")
	}

	fmt.Println()

	if allOk {
		fmt.Println("All required dependencies are available.")
		return nil
	}

	fmt.Println("Some dependencies are missing. Install with:")
	fmt.Println("  npm install -g zod-to-json-schema tsx typescript")
	return fmt.Errorf("missing dependencies")
}
