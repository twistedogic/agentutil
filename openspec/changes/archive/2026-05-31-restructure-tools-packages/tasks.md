## 1. Create new package directories

- [x] 1.1 Create `tools/lsp/` directory
- [x] 1.2 Create `tools/fetch/` directory

## 2. Move and rename LSP tools

- [x] 2.1 Copy `tools/eino/tools.go` to `tools/lsp/tools.go`
- [x] 2.2 Change package declaration in `tools/lsp/tools.go` from `package tools` to `package lsp`

## 3. Move and rename fetch tools

- [x] 3.1 Copy `tools/eino/fetch.go` to `tools/fetch/fetch.go`
- [x] 3.2 Copy `tools/eino/fetch_helpers.go` to `tools/fetch/fetch_helpers.go`
- [x] 3.3 Copy `tools/eino/fetch_helpers_test.go` to `tools/fetch/fetch_helpers_test.go`
- [x] 3.4 Copy `tools/eino/fetch_test.go` to `tools/fetch/fetch_test.go`
- [x] 3.5 Change package declarations in all `tools/fetch/` files from `package tools` to `package fetch`

## 4. Clean up

- [x] 4.1 Delete `tools/eino/` directory and all its contents

## 5. Verify

- [x] 5.1 Run `go build ./...` and confirm no errors
- [x] 5.2 Run `go test ./tools/...` and confirm all tests pass
