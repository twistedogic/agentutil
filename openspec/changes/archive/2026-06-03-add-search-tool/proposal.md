## Why

The CLI exposes fetch and wiki tools for agents but has no web search capability. Adding a `search` command gives agents a DuckDuckGo-backed search tool using the same patterns as the existing `fetch` tool.

## What Changes

- New `tools/search` package with DuckDuckGo scraping logic (ported from charmbracelet/crush)
- New `search.go` CLI entry point wiring up `newSearchCmd()`
- New `search` subcommand registered in `main.go`

## Capabilities

### New Capabilities
- `search`: Web search via DuckDuckGo lite, returning structured JSON results with title, URL, and snippet

### Modified Capabilities
<!-- none -->

## Impact

- New file: `tools/search/search.go`
- New file: `search.go` (CLI command)
- Modified: `main.go` (register `newSearchCmd`)
- Dependencies: `golang.org/x/net/html` (already in go.mod)
