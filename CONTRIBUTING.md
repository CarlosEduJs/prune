# Contributing to Prune

This guide outlines the technical standards and workflows for contributing to the Prune project.

## Development Environment

### Requirements
- Go 1.24 or higher.

### Setup
1. Clone the repository.
2. Run `go mod tidy` to synchronize dependencies.

## Project Architecture

- **cmd/prune**: Entrypoint and CLI command logic.
- **internal/scan**: File system traversal and entrypoint-driven discovery.
- **internal/lang/js**: The core analysis engine. It uses Tree-sitter to parse source files, extract symbols, and map cross-file references.
- **internal/rules**: Rule definitions and filtering logic.
- **internal/config**: Configuration schema and default settings.
- **internal/report**: Logic for formatting findings into tables or JSON.

## Adding Analysis Rules

New rules can be implemented to detect specific patterns of dead or suspicious code.

1. **Registration**: Add the rule ID and description to the `All()` function in `internal/rules/rules.go`.
2. **Symbol Collection**: If the rule requires new data from the source code, modify `internal/lang/js/collector.go` or `internal/lang/js/ast.go` to extract the necessary information.
3. **Logic Implementation**: Update `internal/lang/js/rules.go` to evaluate the collected symbols against the dependency graph.
4. **Defaults**: Add the rule to the default configuration in `internal/config/default.go` if applicable.

## Testing Standards

All changes must include corresponding tests.

- **Unit Tests**: Place tests alongside the implementation (e.g., `internal/lang/js/collector_test.go`).
- **End-to-End Tests**: Add new test scenarios to `internal/cli/e2e_test.go` or create a new project directory in `examples/`.

Run the full test suite:
```bash
go test ./...
```

## Coding Guidelines

- **Formatting**: Run `go fmt ./...` before committing.
- **Linting**: Run `go vet ./...` to check for common mistakes.
- **Error Handling**: Use wrapped errors for context (e.g., `fmt.Errorf("reading file: %w", err)`).
- **Concurrency**: Use `context.Context` for cancellation and timeouts in scanning logic.

## Pull Request Process

1. Create a feature branch from `main`.
2. Ensure the code is self-documented and adheres to existing patterns.
3. Provide a clear description of the problem solved or the feature added.
4. Verify that the scan performance is not significantly impacted on large projects.
