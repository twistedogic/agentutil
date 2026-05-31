## 1. Package scaffold

- [x] 1.1 Create `tools/wiki/` directory
- [x] 1.2 Create `tools/wiki/wiki.go` with `WikiResult` struct, `WikiSearch` function, and `NewWikiTool` AgentTool

## 2. Core implementation

- [x] 2.1 Implement OpenSearch API call: GET `https://en.wikipedia.org/w/api.php?action=opensearch&search=<query>&limit=1` and decode response
- [x] 2.2 Return error `no Wikipedia results for "<query>"` when results array is empty
- [x] 2.3 Call `fetch.FetchURLAndConvert` with the resolved URL and populate `WikiResult`

## 3. CLI wiring

- [x] 3.1 Create `cmd/agentutil/wiki.go` with `newWikiCmd()` following the fetch command pattern
- [x] 3.2 Register `newWikiCmd()` in `cmd/agentutil/main.go`

## 4. Tests

- [x] 4.1 Write unit tests for `WikiSearch` covering success, no-results, and malformed-response cases (use httptest)
- [x] 4.2 Verify `NewWikiTool` returns error response for empty query
