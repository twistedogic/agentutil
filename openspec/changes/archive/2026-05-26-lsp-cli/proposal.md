## Why

There is no way to query LSP diagnostics or references from the command line without writing Go code. A CLI makes the lsp library directly usable in scripts, CI pipelines, and agent tooling without any glue code.

## What Changes

- New `cmd/agentutil` binary with an `lsp` subcommand
- `lsp diagnostics` command: accepts one or more glob patterns, opens matched files, waits for diagnostics to settle, and emits a flat JSON array
- `lsp refs` command: accepts a file, line, and column, and emits a flat JSON array of locations
- Auto-LSP: language server is auto-detected from file extension using built-in defaults; overridable via `--server` flag
- Workspace root auto-detected by walking up from the target file to find a root marker (`go.mod`, `package.json`, `Cargo.toml`, etc.)

## Capabilities

### New Capabilities

- `lsp-cli-diagnostics`: Run LSP diagnostics on files matched by glob patterns and emit results as JSON
- `lsp-cli-refs`: Find all references to a symbol at a given file/line/col and emit results as JSON

### Modified Capabilities

## Impact

- New `cmd/agentutil/` directory with `main.go`
- Depends on existing `lsp` and `config` packages in this module
- New CLI dependency (e.g., `cobra` or `flag`-based) added to `go.mod`
- No changes to existing library API
