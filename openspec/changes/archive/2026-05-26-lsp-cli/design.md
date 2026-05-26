## Context

`agentutil/lsp` is a Go library for managing LSP clients. It has no CLI surface. Developers who want diagnostics or references must write Go code to use it. The goal is a thin CLI binary that exposes the most useful operations — diagnostics and find-references — as JSON-emitting commands.

## Goals / Non-Goals

**Goals:**
- Single binary `agentutil` with an `lsp` subcommand tree
- `lsp diagnostics <glob...>` — open matched files, wait for diagnostics to settle, emit flat JSON array
- `lsp refs <file> <line> <col>` — find all references at a symbol position, emit flat JSON array
- Auto-detect workspace root by walking up from target file(s) to find a root marker
- Auto-detect language server from file extensions using built-in defaults
- `--server` flag to force a specific server by name
- `--timeout` flag (default 30s) to bound LSP startup + settle time
- Exit codes: `0` = success (clean or results found), `1` = tool/runtime error
- Machine-readable JSON output only (no human-formatted output)

**Non-Goals:**
- Interactive / REPL mode
- Daemon / persistent server process
- Human-readable formatted output
- Server lifecycle management (start/stop/status)
- Code actions, hover, completion, or other LSP features beyond diagnostics and refs

## Decisions

### CLI framework: `flag` stdlib vs `cobra`

Use **cobra**. The lsp subcommand tree (`agentutil lsp diagnostics`, `agentutil lsp refs`) maps naturally to cobra's command hierarchy. Cobra is already common in Go CLI tooling and keeps help/usage handling clean. The stdlib `flag` package becomes unwieldy with subcommands.

### Glob expansion: shell vs library

Use **`filepath.Glob`** (or `doublestar` for `**` support) in the binary itself rather than relying on shell expansion. This makes behavior consistent across shells and allows the tool to be called from agents/scripts that don't shell-expand globs. The `doublestar` library supports `**` patterns which are expected by users.

### Workspace root detection

Walk up from the first target file's directory, checking for: `go.mod`, `package.json`, `Cargo.toml`, `pyproject.toml`, `requirements.txt`, `.git`. Stop at the first match. If none found, use the directory of the target file. This matches what the existing Manager already supports via `RootMarkers`.

### Output schema

**Diagnostics** — flat JSON array, each object:
```json
{
  "file": "/abs/path/to/file.go",
  "line": 10,
  "col": 5,
  "severity": "error",
  "message": "undefined: foo",
  "source": "gopls"
}
```

**References** — flat JSON array, each object:
```json
{
  "file": "/abs/path/to/file.go",
  "line": 10,
  "col": 5
}
```

All paths are absolute. `severity` is one of `"error"`, `"warning"`, `"information"`, `"hint"`.

### ConfigStore for CLI

Implement a minimal `cliConfigStore` that satisfies the `lsp.ConfigStore` interface:
- `LSP()` returns a map with an optional single server override (from `--server` flag)
- `AutoLSP()` returns `true` (always enabled in CLI)
- `Resolver()` returns a no-op resolver (no variable substitution needed)

This avoids needing a config file while reusing the existing Manager plumbing.

## Risks / Trade-offs

- **Slow startup on large projects** → Mitigated by `--timeout` flag; default 30s is generous for most projects
- **One-shot pays startup cost every invocation** → Accepted trade-off; use-case is CI/scripting not interactive editing
- **`**` glob patterns need `doublestar` library** → Small extra dependency; without it `**` silently matches nothing which is confusing

## Open Questions

- Should `lsp diagnostics` with zero matched files be an error (exit 1) or return an empty array?
