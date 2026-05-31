## Context

The project has a `lsp/` package at the root alongside `tools/`, `cmd/`, and `config/`. The `tools/` directory already contains `tools/fetch/` and `tools/lsp/` (stub). The `lsp/` package provides LSP client, manager, handlers, types, and versioned-document utilities. All tool-level packages should live under `tools/`.

## Goals / Non-Goals

**Goals:**
- Move `lsp/*.go` files into `tools/lsp/`
- Update import paths from `github.com/twistedogic/agentutil/lsp` → `github.com/twistedogic/agentutil/tools/lsp`
- Ensure the project builds and tests pass after the move

**Non-Goals:**
- No API changes
- No behavioral changes
- No refactoring of the lsp package internals

## Decisions

**Single package rename, no intermediate shim**: Because this is an internal package with no external consumers expected, we do a direct move with import path updates rather than leaving a forwarding shim. This keeps the codebase clean.

## Risks / Trade-offs

- Any external tooling or scripts that hardcode `agentutil/lsp` will break → low risk (internal tool)
- `tools/lsp/` already exists as a stub; must confirm it is empty or merge correctly before moving
