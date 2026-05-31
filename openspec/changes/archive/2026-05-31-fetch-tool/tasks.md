## 1. Dependencies

- [x] 1.1 Add `github.com/PuerkitoBio/goquery` as a direct dependency in `go.mod`
- [x] 1.2 Add `github.com/JohannesKaufmann/html-to-markdown` as a direct dependency in `go.mod`
- [x] 1.3 Run `go mod tidy` to update `go.sum`

## 2. Helpers

- [x] 2.1 Create `tools/eino/fetch_helpers.go` with `BrowserUserAgent` constant
- [x] 2.2 Implement `removeNoisyElements(html string) string` using `golang.org/x/net/html`
- [x] 2.3 Implement `ConvertHTMLToMarkdown(html string) (string, error)` using `html-to-markdown`
- [x] 2.4 Implement `cleanupMarkdown(content string) string` (collapse blank lines, trim trailing whitespace)
- [x] 2.5 Implement `ExtractLinks(doc *goquery.Document, base *url.URL) []string` — collect `a[href]`, resolve to absolute, filter mailto/javascript/#-only hrefs
- [x] 2.6 Implement `FetchURLAndConvert(ctx, client, url) (content string, links []string, err error)` combining fetch + parse + extract

## 3. Tool

- [x] 3.1 Create `tools/eino/fetch.go` defining `FetchResult{Content string, Links []string}`
- [x] 3.2 Implement `NewFetchTool(client *http.Client) fantasy.AgentTool` with tool name `"fetch"` and description
- [x] 3.3 Validate `url` param is non-empty and starts with `http://` or `https://`; return error response otherwise
- [x] 3.4 Call `FetchURLAndConvert` and return a `fantasy.NewTextResponse` with JSON-encoded `FetchResult`

## 4. Tests

- [x] 4.1 Add `tools/eino/fetch_helpers_test.go` — unit tests for `ExtractLinks` (relative, absolute, filtered hrefs)
- [x] 4.2 Add `tools/eino/fetch_test.go` — table-driven tests for `NewFetchTool` using `httptest.NewServer` (success, non-200, empty URL, invalid scheme)
- [x] 4.3 Run `go test ./tools/eino/...` and confirm all pass
