## 1. Core Package

- [x] 1.1 Create `tools/search/search.go` with `SearchResult` type, `Search(ctx, client, query, maxResults)` function, DDG scraping helpers, and pre-request delay
- [x] 1.2 Create `tools/search/search_test.go` with unit tests for HTML parsing (offline) and delay behavior
- [x] 1.3 Create `tools/search/tool.go` with `SearchResponse` wrapper type and `NewSearchTool(client *http.Client) fantasy.AgentTool`

## 2. CLI Wiring

- [x] 2.1 Create `search.go` at module root
- [x] 2.2 Register `newSearchCmd()` in `main.go`

## 3. Verification

- [x] 3.1 Run `go build ./...` and confirm no errors
- [x] 3.2 Run `go test ./tools/search/...`
- [x] 3.3 Smoke-test `agentutil search "golang" --max 3` and confirm JSON output
