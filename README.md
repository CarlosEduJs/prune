# Prune

> [!NOTE]
> **Prune is currently in Beta (v0.1.0-beta.1).** It is under active development, and features, CLI flags, or output formats may change before the stable release.

Static analysis tool designed to identify dead code in JavaScript and TypeScript projects (for now...).

## Description

Prune identifies unreachable code by building a dependency graph from defined entrypoints. It helps maintain clean codebases by detecting files, exports, functions, and variables that are no longer referenced in the application lifecycle.

## Features

- Dependency graph analysis utilizing Tree-sitter grammars.
- Support for `.js`, `.jsx`, `.ts`, and `.tsx` extensions.
- Detection types:
    - Orphaned files (files never imported).
    - Unused exports (symbols exported but never imported).
    - Unused functions and variables.
    - Suspicious dynamic code usage (e.g., `eval`, `Function`).
- Human-friendly CLI output format grouped by file (`pretty`).
- Machine-readable output (JSON, NDJSON) for automation.
- **Path alias resolution** for TypeScript aliases like `@/` and `~/`.
- **Streaming mode** for partial results in real-time.
- CI/CD integration via exit codes and finding thresholds.
- Cross-platform support (Linux, macOS, Windows).

## How it Works

The tool is written in Go and uses Tree-sitter for high-performance parsing. 

- **internal/scan**: Manages file system crawling and glob pattern matching.
- **internal/lang/js**: Implements the JavaScript and TypeScript parsers and AST traversal logic.
- **internal/rules**: Contains the logic to correlate definitions and references to identify dead code.
- **internal/report**: Handles data formatting for the CLI.

The analysis starts from the `entrypoints` defined in the configuration. Any file or symbol not reachable from these roots is reported.

## Installation

### Requirements

- Go 1.24 or higher.

### Building from Source

```bash
go build -o prune ./cmd/prune
```

### Check version

```bash
prune version
```

## Usage

### Initialize

Generate a default configuration file in the current directory:

```bash
prune init
```

### Scan

Perform code analysis:

```bash
prune scan
```

Flags:
- `--config`: Path to the configuration file (default: `prune.yaml`).
- `--format`: Output format: `pretty`, `json`, `ndjson`, or `table` (alias for `pretty`).
- `--min-confidence`: Minimum confidence level to report (`safe`, `likely_dead`, `review`).
- `--fail-on-findings`: Exit with a non-zero status code if problems are detected.
- `--stream`: Enable streaming mode for partial results in real-time.
- `--stream-interval`: Interval in ms between stream flushes (default: 250ms).


Pretty Output

```bash
Prune v0.1.0-beta.1 — 11 issues found in 9ms

✔ SAFE (1)

  utils/unused.ts
  └─ unused file: unused.ts

⚠ REVIEW (10)

  components/Dashboard.tsx
  └─ possible dynamic usage: user.name
  └─ possible dynamic usage: user.email

  main.ts
  └─ possible dynamic usage: console.log
  └─ possible dynamic usage: db.save
  └─ possible dynamic usage: bootstrap().catch
  └─ possible dynamic usage: console.error

  services/db.ts
  └─ possible dynamic usage: this.items.push
  └─ possible dynamic usage: this.items
  └─ possible dynamic usage: this.items.find
  └─ possible dynamic usage: i.id

─────────────────────────────────
Summary

  Files        1
  Dynamic      10

  SAFE         1
  REVIEW       10

  Total        11

Done in 9ms
```

### Streaming Mode

For large projects, you can use streaming mode to receive results incrementally:

```bash
# Stream results as NDJSON (one JSON object per line)
prune scan --stream --format ndjson

# Faster flush interval (100ms)
prune scan --stream --stream-interval 100

# Can also use with json format (auto-converted to ndjson)
prune scan --stream --format json
```

Streaming outputs findings in real-time as files are processed. This is useful for:
- Large codebases where you want immediate feedback
- CI/CD pipelines that want to stream results to external systems
- Debugging analysis progress

### List Rules

List all available analysis rules:

```bash
prune rules
```

## Configuration

Configuration is managed via `prune.yaml`.

```yaml
version: 2
project:
  name: my-project
  language: js-ts

ts_config:
  enabled: true
  baseUrl: .
  paths:
    "@/*":
      - src/*
    "~/*":
      - src/components/*

scan:
  paths:
    - src
  include:
    - "**/*.ts"
    - "**/*.tsx"
  exclude:
    - "node_modules/**"
    - "dist/**"
  stream:
    enabled: true
    interval_ms: 250
entrypoints:
  files:
    - src/main.ts
rules:
  unused_function:
    enabled: true
  unused_export:
    enabled: true
report:
  format: pretty
  min_confidence: safe
```

## Development

### Running Tests

```bash
go test ./...
```

### Manual Testing

The project includes examples in the `examples/` directory:

```bash
go run ./cmd/prune scan --config examples/js-complex/prune.yaml
```

## CI/CD Integration

To utilize Prune in a CI/CD environment, you can install the binary and incorporate it as a validation step.

### GitHub Actions Example

```yaml
name: Code Quality
on: [push, pull_request]

jobs:
  analysis:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.24'
          
      - name: Install Prune
        run: go install github.com/carlosedujs/prune/cmd/prune@latest
        
      - name: Run Analysis
        run: prune scan --fail-on-findings
```

The `--fail-on-findings` flag ensures that the pipeline exits with a non-zero status code if unreachable dead code is detected, facilitating automated code quality enforcement.

## Limitations / Known Issues

> [!TIP]
> **Handling False Positives:** Dynamic code patterns (like `eval`, bracket notation `obj[var]`, or dynamic imports) can't always be strictly resolved statically. Prune flags these as `REVIEW` rather than `SAFE` so you can manually verify them, minimizing the risk of accidentally deleting actively used code!

- Support is currently restricted to JavaScript and TypeScript ecosystems.
- Dynamic imports or class property access via strings may result in false positives (marked as `REVIEW`).
- Large codebases with circular dependencies may require higher memory allocation during Graph traversal.
- Scoped alias patterns like `@scope/*` are not supported.

## License

MIT
