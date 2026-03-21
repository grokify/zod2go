#!/usr/bin/env node
/**
 * zod-extract.mjs
 *
 * Extracts a Zod schema from a TypeScript file and converts it to JSON Schema.
 *
 * Usage:
 *   node zod-extract.mjs <input.ts> <exportName> [--output <file.json>]
 *
 * Example:
 *   node zod-extract.mjs schema.ts MySchema > schema.json
 *   node zod-extract.mjs schema.ts MySchema --output schema.json
 */

import { execSync } from 'child_process';
import { existsSync, readFileSync, writeFileSync, mkdtempSync } from 'fs';
import { join, dirname, basename, resolve } from 'path';
import { tmpdir } from 'os';
import { fileURLToPath } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

function printUsage() {
  console.error(`
Usage: zod-extract.mjs <input.ts> <exportName> [options]

Arguments:
  input.ts     TypeScript file containing Zod schema
  exportName   Name of the exported Zod schema to convert

Options:
  --output, -o    Output file (default: stdout)
  --help, -h      Show this help message

Example:
  node zod-extract.mjs schema.ts UserSchema > user.schema.json
  node zod-extract.mjs ./types.ts ConfigSchema -o config.schema.json
`);
}

function parseArgs(args) {
  const result = {
    input: null,
    exportName: null,
    output: null,
  };

  let i = 0;
  while (i < args.length) {
    const arg = args[i];

    if (arg === '--help' || arg === '-h') {
      printUsage();
      process.exit(0);
    } else if (arg === '--output' || arg === '-o') {
      result.output = args[++i];
    } else if (!result.input) {
      result.input = arg;
    } else if (!result.exportName) {
      result.exportName = arg;
    }
    i++;
  }

  return result;
}

async function main() {
  const args = parseArgs(process.argv.slice(2));

  if (!args.input || !args.exportName) {
    console.error('Error: Missing required arguments');
    printUsage();
    process.exit(1);
  }

  const inputPath = resolve(args.input);

  if (!existsSync(inputPath)) {
    console.error(`Error: Input file not found: ${inputPath}`);
    process.exit(1);
  }

  // Create a temporary directory for the conversion
  const tempDir = mkdtempSync(join(tmpdir(), 'zod2go-'));
  const wrapperPath = join(tempDir, 'wrapper.mjs');

  // Create a wrapper script that imports the schema and converts it
  const wrapperContent = `
import { zodToJsonSchema } from 'zod-to-json-schema';
import { ${args.exportName} } from '${inputPath}';

const jsonSchema = zodToJsonSchema(${args.exportName}, {
  name: '${args.exportName}',
  $refStrategy: 'none',
  errorMessages: true,
});

console.log(JSON.stringify(jsonSchema, null, 2));
`;

  writeFileSync(wrapperPath, wrapperContent);

  try {
    // Check if zod-to-json-schema is installed
    try {
      execSync('npm list zod-to-json-schema', { stdio: 'pipe', cwd: tempDir });
    } catch {
      // Install dependencies in temp directory
      console.error('Installing dependencies...');
      execSync('npm init -y && npm install zod zod-to-json-schema typescript ts-node', {
        cwd: tempDir,
        stdio: 'pipe',
      });
    }

    // Run the wrapper with ts-node or node
    let output;
    try {
      // Try with tsx first (handles TypeScript imports)
      output = execSync(`npx tsx "${wrapperPath}"`, {
        cwd: dirname(inputPath),
        encoding: 'utf-8',
        stdio: ['pipe', 'pipe', 'pipe'],
      });
    } catch {
      // Fall back to node with experimental modules
      output = execSync(`node --experimental-specifier-resolution=node "${wrapperPath}"`, {
        cwd: dirname(inputPath),
        encoding: 'utf-8',
        stdio: ['pipe', 'pipe', 'pipe'],
      });
    }

    if (args.output) {
      writeFileSync(args.output, output);
      console.error(`Written to: ${args.output}`);
    } else {
      console.log(output);
    }

  } catch (error) {
    console.error('Error converting schema:', error.message);
    if (error.stderr) {
      console.error(error.stderr.toString());
    }
    process.exit(1);
  } finally {
    // Cleanup temp directory
    try {
      execSync(`rm -rf "${tempDir}"`, { stdio: 'pipe' });
    } catch {
      // Ignore cleanup errors
    }
  }
}

main().catch(err => {
  console.error('Fatal error:', err);
  process.exit(1);
});
