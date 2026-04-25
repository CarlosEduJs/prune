# Prune

> [!NOTE]
> **Prune is currently in Beta (`v0.4.0-beta.1`).** It is under active development, and features, CLI flags, or output formats may change before stable release.

Prune is a static analysis CLI that finds dead code in JavaScript and TypeScript projects by building a reachability graph from your configured entrypoints.

## Features

- Tree-sitter-based analysis for `.js`, `.jsx`, `.ts`, and `.tsx`
- Finds:
  - `unused_file`
  - `unused_export`
  - `unused_function`
  - `unused_variable`
  - `possible_dynamic_usage`
  - `suspicious_dynamic_usage`
- Confidence levels: `safe`, `likely_dead`, `review`
- Output formats: `pretty` (and `table` alias), `json`, `ndjson`
- Streaming mode with configurable interval and batch size
- TypeScript path alias support via `ts_config` (`@/*`, `~/`, `@scope/*`, etc.)
- CI/CD-ready via `--fail-on-findings`

## Installation

### Requirements

- Go `1.24+`
- CGO enabled (required by Tree-sitter bindings)

### Build from source

```bash
go build -o prune ./cmd/prune
```

### Install with `go install`

```bash
go install github.com/carlosedujs/prune/cmd/prune@latest
```

### Verify

```bash
prune version
```

Expected output:

```text
prune version  0.4.0-beta.1
```

## Quick Start

Initialize config:

```bash
prune init
```

Run scan:

```bash
prune scan
```

Common scan flags:

- `--config`: Path to config (default: `prune.yaml`)
- `--format`: `pretty`, `json`, `ndjson`, or `table`
- `--min-confidence`: `safe`, `likely_dead`, `review`
- `--paths` (repeatable): Override scan paths
- `--fail-on-findings`: Exit non-zero if findings remain after filters
- `--stream`: Enable streaming mode
- `--stream-interval`: Flush interval in milliseconds (default: `250`)
- `--stream-batch-size`: Files per stream batch (default: `50`)
- `--compact`: Show summary counts only
- `--only`: Show only one confidence level
- `--deletable`: Show only `unused_file` findings with `safe` confidence

## Output

### Pretty (default)

```text
Prune v0.4.0-beta.1 — 4 issues found in 340ms

✔ SAFE (2)

  src/utils/legacy.ts
  └─ unused file: legacy.ts

✖ LIKELY DEAD (1)

  src/lib/helpers.ts
  └─ unused function: formatDeprecated

⚠ REVIEW (1)

  src/loader.ts
  └─ suspicious dynamic usage: require

─────────────────────────────────
Summary
─────────────────────────────────
  Files        1
  Functions    1
  Suspicious   1

  SAFE         2
  LIKELY DEAD  1
  REVIEW       1

  Total        4
─────────────────────────────────
Done in 340ms
```

### JSON

```bash
prune scan --format json
```

Returns a root object with:

- `summary`
- `findings`
- `metadata`

### NDJSON

```bash
prune scan --format ndjson
```

Streaming mode is designed for NDJSON:

```bash
prune scan --stream --format ndjson
prune scan --stream --stream-interval 100 --stream-batch-size 20 --format ndjson
```

If `--stream` is used with `--format json`, output is promoted to NDJSON.

## Configuration (`prune.yaml`)

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
    - "**/*.js"
    - "**/*.jsx"
  exclude:
    - node_modules/**
    - dist/**
    - build/**
  stream:
    enabled: false
    interval_ms: 250
    batch_size: 50

entrypoints:
  files:
    - src/main.ts
  patterns:
    - src/pages/**

rules:
  unused_function:
    enabled: true
    confidence:
      default: likely_dead
      if_high_risk_dynamic: review
  unused_variable:
    enabled: true
    confidence:
      default: safe
      if_exported: likely_dead
      if_high_risk_dynamic: review
  unused_export:
    enabled: true
  unused_file:
    enabled: true
  possible_dynamic_usage:
    enabled: true
  suspicious_dynamic_usage:
    enabled: true

report:
  format: table
  min_confidence: safe
```

## CI/CD

GitHub Actions example:

```yaml
name: Code Quality
on: [push, pull_request]

jobs:
  dead-code:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: "1.24"

      - name: Install Prune
        run: go install github.com/carlosedujs/prune/cmd/prune@latest

      - name: Run analysis
        run: prune scan --fail-on-findings --min-confidence safe
```

## Development

Run tests:

```bash
go test ./...
```

Run against example project:

```bash
go run ./cmd/prune scan --config examples/js-complex/prune.yaml
```

## Limitations

- Static analysis cannot fully resolve all runtime-dynamic patterns.
- External consumers outside scan scope can make exports look unused.
- Syntax-error files fall back to regex extraction, which is less precise.

See full docs in `docs/content/docs/limitations.mdx`.

## License

MIT
