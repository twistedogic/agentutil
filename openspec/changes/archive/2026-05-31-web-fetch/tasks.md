## 1. CLI Command

- [x] 1.1 Create `cmd/agentutil/fetch.go` with `newFetchCmd()` — cobra command accepting one URL arg and `--timeout` flag, calling `fetch.FetchURLAndConvert`, marshalling `FetchResult` to stdout via `writeJSON`
- [x] 1.2 Register `newFetchCmd()` in `cmd/agentutil/main.go`

## 2. Skill

- [x] 2.1 Create `skills/web-fetch/SKILL.md` documenting `agentutil fetch <url>`, output shape, common patterns (crawl via links), and edge cases (non-HTML, large pages)

## 3. Verification

- [x] 3.1 Run `go build ./cmd/agentutil/...` — verify no compile errors
- [x] 3.2 Run `agentutil fetch https://example.com` — verify JSON output with `content` and `links`
