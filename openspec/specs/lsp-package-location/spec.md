## ADDED Requirements

### Requirement: LSP package lives under tools/
The `lsp` package SHALL be located at `tools/lsp/` within the module, consistent with other tool-level packages.

#### Scenario: Import path uses tools/lsp
- **WHEN** any Go file imports the LSP package
- **THEN** the import path SHALL be `github.com/twistedogic/agentutil/tools/lsp`
