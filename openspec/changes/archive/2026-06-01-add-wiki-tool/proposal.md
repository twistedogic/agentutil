## Why

Agents frequently need encyclopedic reference content. The existing `fetch` tool requires knowing the exact URL; a `wiki` tool lets agents search Wikipedia by topic and get article content in one call, without needing to construct or know URLs.

## What Changes

- New `tools/wiki` package exposing a `WikiSearch` function and a `NewWikiTool` fantasy.AgentTool
- New `agentutil wiki` CLI command that takes a search query and outputs JSON with title, URL, content, and links
- The tool uses the Wikipedia OpenSearch API to resolve a query to a canonical article URL, then delegates to the existing `fetch` plumbing to retrieve and convert the article

## Capabilities

### New Capabilities
- `wiki`: Search Wikipedia by query and return the top article's content as markdown, including title, canonical URL, and extracted links

### Modified Capabilities

## Impact

- New package `tools/wiki/` depending on `tools/fetch`
- New CLI subcommand registered in `cmd/agentutil/main.go`
- No breaking changes; purely additive
