## Why

The `lsp/` package lives at the project root while other tool packages live under `tools/`. Moving it to `tools/lsp/` aligns the project structure so all tool-level packages are co-located.

## What Changes

- Move all files from `lsp/` to `tools/lsp/`
- Update Go package import paths from `github.com/...agentutil/lsp` to `github.com/...agentutil/tools/lsp`
- Remove the now-empty `lsp/` directory

## Capabilities

### New Capabilities
<!-- None - this is a structural refactor only -->

### Modified Capabilities
<!-- No spec-level behavior changes; this is purely a package relocation -->

## Impact

- `lsp/` package files moved to `tools/lsp/`
- All callers importing `agentutil/lsp` must update their import path to `agentutil/tools/lsp`
- No API or behavioral changes
