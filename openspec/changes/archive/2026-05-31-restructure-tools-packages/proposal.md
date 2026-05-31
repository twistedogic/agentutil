## Why

`tools/eino` is a poorly named catch-all that mixes LSP tool wrappers with HTTP fetch tools. The name reflects an old framework dependency rather than the domain. As the tools package grows, co-locating unrelated tools in a single directory makes navigation and ownership unclear.

## What Changes

- **Rename** `tools/eino/` → split into two focused packages:
  - `tools/lsp/` — LSP tool wrappers (`DiagnosticsTool`, `ReferencesTool`, `RestartTool`)
  - `tools/fetch/` — HTTP fetch tool (`NewFetchTool`, `FetchURLAndConvert`, `ExtractLinks`, helpers)
- **Rename package declarations**: `package tools` → `package lsp` in `tools/lsp/`, `package fetch` in `tools/fetch/`
- **Delete** `tools/eino/` directory

No behavior changes — this is a pure structural refactor.

## Capabilities

### New Capabilities

### Modified Capabilities

## Impact

- Files moved: `tools/eino/tools.go` → `tools/lsp/tools.go`, `tools/eino/fetch*.go` → `tools/fetch/`
- Package import paths change: `github.com/twistedogic/agentutil/tools/eino` → `github.com/twistedogic/agentutil/tools/lsp` and `github.com/twistedogic/agentutil/tools/fetch`
- No current consumers of `tools/eino` — zero import paths to update outside the package itself
- No dependency changes
