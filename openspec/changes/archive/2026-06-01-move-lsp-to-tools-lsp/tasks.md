## 1. Move Package Files

- [x] 1.1 Verify `tools/lsp/` contents and ensure no conflicts with `lsp/` files
- [x] 1.2 Move all `.go` files from `lsp/` to `tools/lsp/` (client.go, handlers.go, lsp_test.go, manager.go, types.go, versioned.go)
- [x] 1.3 Update the `package` declaration in each moved file if needed (confirm package name stays `lsp`)
- [x] 1.4 Remove the now-empty `lsp/` directory

## 2. Update Import Paths

- [x] 2.1 Update `cmd/agentutil/lsp.go` import from `github.com/twistedogic/agentutil/lsp` to `github.com/twistedogic/agentutil/tools/lsp`
- [x] 2.2 Update `tools/lsp/tools.go` self-import if it references the old path

## 3. Verify

- [x] 3.1 Run `go build ./...` and confirm no errors
- [x] 3.2 Run `go test ./...` and confirm all tests pass
