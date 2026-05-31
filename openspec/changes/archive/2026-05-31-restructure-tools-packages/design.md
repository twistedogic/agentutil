## Context

`tools/eino/` currently contains two unrelated sets of tools under a single `package tools` declaration: LSP wrappers (`tools.go`) and HTTP fetch tools (`fetch.go`, `fetch_helpers.go`). The directory is named after a framework rather than a domain. No code outside the package currently imports it.

## Goals / Non-Goals

**Goals:**
- Split `tools/eino/` into `tools/lsp/` (package `lsp`) and `tools/fetch/` (package `fetch`)
- Rename package declarations to match directory names per Go convention
- Delete the `tools/eino/` directory

**Non-Goals:**
- Changing any function signatures or behavior
- Moving the `lsp/` root package (LSP client/manager — separate from tool wrappers)
- Adding new tools

## Decisions

**`tools/lsp` not `lsp/tools`**
The `lsp/` root package already owns the LSP client and manager. Putting tool wrappers under `tools/lsp/` keeps the `tools/` subtree as the home for `fantasy.AgentTool` adapters, and avoids adding to the `lsp/` package which has a different responsibility.

**Package name = directory name**
`package lsp` in `tools/lsp/` and `package fetch` in `tools/fetch/` follows standard Go convention and makes import usage readable: `lsp.DiagnosticsTool(...)`, `fetch.NewFetchTool(...)`.

## Risks / Trade-offs

- [Flat `tools/` grows unbounded] → Each new tool domain gets its own subdirectory; the pattern scales cleanly.
- [No consumers today] → Zero migration cost now; deferring would only increase it.
