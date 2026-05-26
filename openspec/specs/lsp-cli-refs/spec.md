## ADDED Requirements

### Requirement: Refs command finds all references to a symbol
The CLI SHALL accept a file path, line number (1-based), and column number (1-based) as positional arguments to the `agentutil lsp refs` command, and emit all references to the symbol at that position as a flat JSON array.

#### Scenario: References found
- **WHEN** user runs `agentutil lsp refs ./pkg/foo.go 42 10`
- **THEN** all references to the symbol at line 42, col 10 are emitted as a JSON array

#### Scenario: No references found
- **WHEN** the symbol at the given position has no references
- **THEN** the CLI SHALL emit `[]` and exit with code `0`

### Requirement: Refs output is a flat JSON array
The CLI SHALL emit references as a flat JSON array on stdout. Each element SHALL contain `file` (absolute path), `line` (1-based integer), and `col` (1-based integer).

#### Scenario: Multiple references across files
- **WHEN** symbol is referenced in multiple files
- **THEN** all reference locations are included in the output array with correct `file`, `line`, and `col` fields

### Requirement: Refs command shares workspace root and server detection
The `lsp refs` command SHALL use the same workspace root auto-detection and `--server` / `--timeout` flags as `lsp diagnostics`.

#### Scenario: Workspace root detected from target file
- **WHEN** user runs `agentutil lsp refs ./pkg/foo/foo.go 10 5`
- **THEN** workspace root is detected by walking up from `./pkg/foo/`

#### Scenario: --timeout applied to refs
- **WHEN** user provides `--timeout 45s`
- **THEN** LSP startup and indexing is bounded to 45 seconds
