## Context

`agentutil` provides agent-facing CLI tools and fantasy.AgentTools. The `fetch` package (`tools/fetch`) already handles HTTP fetching, HTML→markdown conversion, and link extraction. The `wiki` tool is a thin layer on top: it resolves a free-text query to a canonical Wikipedia URL via the OpenSearch API, then delegates to `fetch` for content retrieval.

## Goals / Non-Goals

**Goals:**
- Resolve a search query to a Wikipedia article URL using the OpenSearch API
- Fetch and convert the top result to markdown via existing `fetch` plumbing
- Expose as a `fantasy.AgentTool` (primary use case) and CLI subcommand
- Return structured JSON: `{title, url, content, links}`
- Return a clear error when the search yields no results

**Non-Goals:**
- Returning multiple results or letting the caller choose
- Caching or offline support
- Languages other than English
- Disambiguation pages (first result is accepted as-is)

## Decisions

**1. Package structure: `tools/wiki/` mirroring `tools/fetch/`**
Consistent with the existing pattern. `tools/wiki/wiki.go` exports `WikiSearch` and `NewWikiTool`; `cmd/agentutil/wiki.go` wires the CLI command.

**2. OpenSearch API over MediaWiki parse API**
`https://en.wikipedia.org/w/api.php?action=opensearch&search=<query>&limit=1` returns a simple JSON array `[query, titles[], descs[], urls[]]`. No auth, no parsing complexity. First URL is used directly.

**3. Reuse `fetch.FetchURLAndConvert` — do not inline**
The fetch package already handles HTML→markdown, link extraction, size limits, and User-Agent headers. No reason to duplicate. `tools/wiki` imports `tools/fetch`.

**4. Error on zero results**
Agents need actionable signals. An empty result set returns `fmt.Errorf("no Wikipedia results for %q", query)`.

**5. No `--limit` flag on CLI**
The tool is agent-oriented (Option B: search + auto-fetch top result). A limit flag adds surface area with no agent benefit.

## Risks / Trade-offs

- [Wikipedia OpenSearch may return unrelated first result for ambiguous queries] → Agents can retry with a more specific query; no mitigation needed in the tool itself
- [Wikipedia HTML is verbose; converted markdown can be large] → Inherited from `fetch`; existing 5MB cap applies
- [OpenSearch API shape could change] → Low risk; stable public API. Decode into `[]json.RawMessage` and validate length to fail clearly.

## Migration Plan

Purely additive. No existing behavior changes. New package + new CLI subcommand registered in `main.go`.
