## 1. Project Setup

- [x] 1.1 Create `cmd/agentutil/main.go` with cobra root command
- [x] 1.2 Add `cobra` and `doublestar` dependencies to `go.mod`
- [x] 1.3 Create `cmd/agentutil/lsp.go` with `lsp` subcommand registered on root

## 2. Shared CLI Plumbing

- [x] 2.1 Implement `findWorkspaceRoot(path string) string` — walk up checking root markers
- [x] 2.2 Implement `cliConfigStore` satisfying `lsp.ConfigStore` — auto-LSP enabled, optional server override, no-op resolver
- [x] 2.3 Add shared `--server` and `--timeout` flags to the `lsp` subcommand (persistent flags inherited by subcommands)

## 3. Diagnostics Command

- [x] 3.1 Create `lsp diagnostics` subcommand in `cmd/agentutil/lsp.go`
- [x] 3.2 Implement glob expansion using `doublestar.Glob` for each positional pattern argument
- [x] 3.3 Deduplicate and collect matched file paths
- [x] 3.4 Start LSP Manager, open all matched files, call `WaitForDiagnostics` with timeout
- [x] 3.5 Collect diagnostics from all open files, map to output struct with `file`, `line`, `col`, `severity`, `message`, `source`
- [x] 3.6 Emit flat JSON array to stdout; emit `[]` when no diagnostics

## 4. Refs Command

- [x] 4.1 Create `lsp refs <file> <line> <col>` subcommand in `cmd/agentutil/lsp.go`
- [x] 4.2 Open the target file, wait for LSP ready, call `FindReferences`
- [x] 4.3 Map `protocol.Location` results to output struct with `file`, `line`, `col`
- [x] 4.4 Emit flat JSON array to stdout; emit `[]` when no references

## 5. Error Handling & Exit Codes

- [x] 5.1 Any runtime/tool error (bad args, LSP failed to start, file not found) writes to stderr as JSON `{"error": "..."}` and exits with code `1`
- [x] 5.2 Successful runs (including empty results) exit with code `0`

## 6. Tests

- [x] 6.1 Unit test `findWorkspaceRoot` with temp directory trees
- [x] 6.2 Unit test glob expansion and deduplication logic
- [x] 6.3 Unit test JSON output marshalling for diagnostics and refs structs
