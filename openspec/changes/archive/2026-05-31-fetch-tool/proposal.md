## Why

agentutil has LSP tools but no way for agents to retrieve content from the web. A `fetch` tool that returns page content as markdown plus all extracted hyperlinks gives agents a useful building block for browsing, crawling, and research tasks.

## What Changes

- Add `FetchTool` — a `fantasy.AgentTool` that takes a URL, fetches the page, converts HTML to markdown, and extracts all `<a href>` links (resolved to absolute URLs).
- Add `fetch_helpers.go` — shared HTTP, HTML→markdown conversion, and link extraction helpers.
- Add `goquery` and `html-to-markdown` as direct dependencies.

## Capabilities

### New Capabilities

- `fetch`: HTTP fetch of a URL returning cleaned markdown content and a list of resolved absolute links extracted from the page HTML.

### Modified Capabilities

## Impact

- New file: `tools/eino/fetch.go`
- New file: `tools/eino/fetch_helpers.go`
- `go.mod` / `go.sum`: adds `github.com/PuerkitoBio/goquery` and `github.com/JohannesKaufmann/html-to-markdown`
- No breaking changes to existing tools.
