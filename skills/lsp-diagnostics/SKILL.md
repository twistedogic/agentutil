---
name: lsp-diagnostics
description: Use when finding code errors, warnings, or quality issues in source files. Use when checking for linting problems, type errors, or compilation issues. Use when reviewing code before commits, PRs, or deployments. Use when debugging mysterious behavior and suspecting hidden errors. Triggered by mentions of errors, warnings, diagnostics, lint, type check, or code quality.
license: MIT
compatibility: Requires agentutil CLI with LSP servers installed (gopls, rust-analyzer, typescript-language-server, clangd, etc.)
metadata:
  author: agentutil
  version: "1.0"
---

# LSP Diagnostics

Get errors, warnings, and other diagnostics from Language Server Protocol (LSP) servers for code files.

## Overview

LSP servers (gopls, rust-analyzer, typescript-language-server, etc.) continuously analyze code and surface issues. This skill helps you query those diagnostics programmatically using the agentutil CLI.

## When to Use

- Checking files for errors before editing
- Finding all issues in a file or directory
- Verifying code is clean before commits/PRs
- Debugging when code "should work" but doesn't
- Finding type errors, lint warnings, or compiler complaints

## Tool Usage

```bash
agentutil lsp diagnostics <pattern> [pattern...]
```

**Arguments:**
- `pattern`: Glob pattern(s) for files to check (e.g., `*.go`, `src/**/*.ts`)

**Options:**
- `--server <name>`: Force specific LSP server (e.g., `gopls`)
- `--timeout <duration>`: Wait time for diagnostics (default: 30s)

## Output Format

```json
[
  {
    "file": "/path/to/file.go",
    "line": 42,
    "col": 5,
    "severity": "error",
    "message": "undefined: someFunc",
    "source": "gopls"
  }
]
```

**Severity levels:** `error`, `warning`, `information`, `hint`

## Common Patterns

### Check single file
```bash
agentutil lsp diagnostics src/main.go
```

### Check all files in directory
```bash
agentutil lsp diagnostics src/
```

### Check with glob pattern
```bash
agentutil lsp diagnostics "**/*.go"
```

### Check for specific LSP server
```bash
agentutil lsp diagnostics --server gopls "*.go"
```

### Increase timeout for large projects
```bash
agentutil lsp diagnostics --timeout 60s "**/*.go"
```

## Technique: Systematic Diagnostics Investigation

When you encounter unexpected behavior:

1. **Check diagnostics first** — Run `agentutil lsp diagnostics <file>` to surface any errors the LSP knows about
2. **Check entire project** — Errors in other files often cause cascading issues
3. **Filter by severity** — Focus on `error` level first, then `warning`
4. **Note the source** — Different LSPs format messages differently (gopls vs rust-analyzer vs tsc)

## Interpreting Diagnostics

| Source | Message Style | Notes |
|--------|--------------|-------|
| gopls | "undefined: foo" | Go compiler-style messages |
| rust-analyzer | "cannot find `foo` in this scope" | Rust compiler-style |
| typescript-language-server | "Property 'foo' does not exist on type 'Bar'" | TypeScript style |
| clangd | "use of undeclared identifier 'foo'" | C/C++ style |

## Common Issues and Solutions

**Empty output but code seems wrong:**
- The LSP might not have finished analyzing — try `--timeout 60s`
- File might not be in a workspace root — ensure you're in a directory with root markers (go.mod, package.json, etc.)

**No server found:**
- LSP server not installed — check with `which gopls` or `which rust-analyzer`
- File type not supported — verify the LSP handles that file extension

**Stale diagnostics:**
- Restart the LSP server: `agentutil lsp restart <server-name>`
- Or restart all: `agentutil lsp restart`

## Workflow Integration

```
Before editing file:
  1. Run diagnostics
  2. Fix errors first
  3. Then make your changes

Before commit:
  1. Run diagnostics on changed files
  2. Verify no new errors introduced
  3. Fix or acknowledge warnings
```

## Installation

### From source
```bash
go install github.com/twistedogic/agentutil/cmd/agentutil@latest
```

### Verify installation
```bash
agentutil lsp --help
```

### Prerequisites

LSP servers must be installed separately. Common servers:

| Language | Server | Install |
|----------|--------|---------|
| Go | gopls | `go install golang.org/x/tools/gopls@latest` |
| Rust | rust-analyzer | `rustup component add rust-analyzer` |
| TypeScript/JS | typescript-language-server | `npm install -g typescript-language-server` |
| C/C++ | clangd | `brew install clangd` or package manager |
| Python | jedi-language-server | `pip install jedi-language-server` |

Verify servers are available:
```bash
which gopls rust-analyzer typescript-language-server clangd
```

## Limitations

- Diagnostics are as good as the LSP server — some servers are faster/more thorough than others
- Large projects may need longer timeouts
- Some LSP servers don't publish all diagnostics (e.g., they may skip "information" level)