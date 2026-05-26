## ADDED Requirements

### Requirement: Diagnostics command accepts glob patterns
The CLI SHALL accept one or more glob patterns as positional arguments to the `lsp diagnostics` command. Patterns SHALL support `**` for recursive matching. The CLI SHALL expand all patterns and collect the union of matched files before opening them.

#### Scenario: Single glob pattern matches files
- **WHEN** user runs `agentutil lsp diagnostics "./src/**/*.go"`
- **THEN** all `.go` files under `./src/` are collected and opened for analysis

#### Scenario: Multiple glob patterns accepted
- **WHEN** user runs `agentutil lsp diagnostics "*.go" "pkg/**/*.go"`
- **THEN** all files matched by either pattern are collected (deduplicated) and opened

#### Scenario: No files matched by pattern
- **WHEN** a glob pattern matches zero files
- **THEN** the CLI SHALL emit an empty JSON array `[]` and exit with code `0`

### Requirement: Diagnostics output is a flat JSON array
The CLI SHALL emit diagnostics as a flat JSON array on stdout. Each element SHALL contain `file` (absolute path), `line` (1-based integer), `col` (1-based integer), `severity` (one of `"error"`, `"warning"`, `"information"`, `"hint"`), `message` (string), and `source` (string, the LSP server name).

#### Scenario: Files have diagnostics
- **WHEN** LSP diagnostics are found across multiple files
- **THEN** all diagnostics are emitted as a single flat JSON array, each with `file`, `line`, `col`, `severity`, `message`, and `source` fields

#### Scenario: No diagnostics found
- **WHEN** LSP reports no diagnostics for any matched file
- **THEN** the CLI SHALL emit `[]` and exit with code `0`

### Requirement: Workspace root is auto-detected
The CLI SHALL determine the workspace root by walking up from the directory of the first matched file, checking for root markers: `go.mod`, `package.json`, `Cargo.toml`, `pyproject.toml`, `requirements.txt`, `.git`. The first directory containing any marker SHALL be used as the workspace root. If no marker is found, the directory of the first matched file SHALL be used.

#### Scenario: go.mod found in ancestor directory
- **WHEN** target file is `/project/pkg/foo/foo.go` and `/project/go.mod` exists
- **THEN** workspace root is set to `/project`

#### Scenario: No root marker found
- **WHEN** no root marker exists in any ancestor directory
- **THEN** workspace root is set to the directory containing the first matched file

### Requirement: Language server is auto-detected
The CLI SHALL automatically select the appropriate language server based on file extensions using built-in defaults. The user MAY override with `--server <name>` to force a specific server.

#### Scenario: Go files without --server flag
- **WHEN** matched files include `.go` files and no `--server` flag is provided
- **THEN** `gopls` is used as the language server

#### Scenario: --server flag overrides auto-detection
- **WHEN** user provides `--server rust-analyzer`
- **THEN** `rust-analyzer` is used regardless of file extensions

### Requirement: Timeout is configurable
The CLI SHALL support a `--timeout` flag (default `30s`) that bounds the total time allowed for LSP startup and diagnostic settling. If the timeout is exceeded, the CLI SHALL emit whatever diagnostics have been collected so far and exit with code `0`.

#### Scenario: Default timeout used
- **WHEN** `--timeout` flag is not provided
- **THEN** a 30-second timeout is applied

#### Scenario: Custom timeout provided
- **WHEN** user runs `agentutil lsp diagnostics --timeout 60s "*.go"`
- **THEN** up to 60 seconds is allowed for LSP startup and diagnostic settling
