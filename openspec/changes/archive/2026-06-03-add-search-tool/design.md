## Context

The project is a CLI (`agentutil`) that exposes tools for AI agents: `fetch`, `wiki`, `lsp`. Each tool follows a pattern: a `tools/<name>/` package with core logic + an agent tool factory, and a `<name>.go` file in the root wiring a cobra command.

The DuckDuckGo search logic is ported from [charmbracelet/crush](https://github.com/charmbracelet/crush/blob/main/internal/agent/tools/search.go), which scrapes `lite.duckduckgo.com` (no API key required).

## Goals / Non-Goals

**Goals:**
- Port DuckDuckGo search logic into `tools/search/search.go`
- Expose `agentutil search <query>` CLI command with `--timeout` and `--max`/`-n` flags
- Return `{"results": [...]}` JSON wrapper matching project conventions
- Apply the upstream rate-limit delay on every invocation (one-shot, no mutex needed in CLI context)
- Expose a `fantasy.AgentTool` following the same pattern as `tools/fetch`

**Non-Goals:**
- Supporting other search engines
- Pagination / cursor-based results
- Caching

## Decisions

### Output shape: `{"results": [...]}`
The project returns structured JSON from all commands. A wrapper object is consistent with `FetchResult` and leaves room for metadata fields later.

### Rate-limit delay: simplified
The upstream uses a mutex + wall-clock check suitable for in-process repeated calls. In a CLI (one process, one invocation) we just sleep a random 500–2000ms before the request—no mutex needed. Same UX outcome, simpler code.

### Package structure: `tools/search/`
Matches existing `tools/fetch/` and `tools/lsp/` layout. Core logic is separately testable. Root `search.go` just wires cobra → package.

### No new dependencies
`golang.org/x/net/html` is already in go.mod. No additional deps.

## Risks / Trade-offs

- [DDG scraping fragility] → Scraping `lite.duckduckgo.com` can break if DDG changes their HTML. Mitigation: isolated parser in `tools/search/` makes it easy to update.
- [Rate limiting / blocking] → DDG may return non-200 or empty results under load. Mitigation: return error with status code, let caller retry.

## Open Questions

- None.
