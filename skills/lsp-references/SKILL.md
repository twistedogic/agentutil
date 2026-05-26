---
name: lsp-references
description: Use when finding where a symbol (function, variable, type, class) is defined or used. Use when tracing code flow, finding all call sites, or understanding usage patterns. Use when renaming or refactoring and need to find all occurrences. Use when investigating bugs and suspecting a symbol is used incorrectly. Triggered by mentions of references, find usages, where is X defined, who calls this, or trace symbol.
license: MIT
compatibility: Requires agentutil CLI with LSP servers installed (gopls, rust-analyzer, typescript-language-server, clangd, etc.)
metadata:
  author: agentutil
  version: "1.0"
---

# LSP References

Find all references to a symbol at a specific position in a file using Language Server Protocol (LSP).

## Overview

LSP servers maintain a full symbol index of your codebase. This skill lets you query that index to find where symbols are defined and used — essential for understanding code flow, refactoring, and debugging.

## When to Use

- Finding all places a function is called
- Finding all places a variable is used
- Finding the definition of a symbol
- Understanding where a type is instantiated
- Refactoring safely (rename, extract, move)
- Investigating unexpected behavior (what's calling this?)

## Tool Usage

```bash
agentutil lsp refs <file> <line> <col>
```

**Arguments:**
- `file`: Path to the source file
- `line`: Line number (1-based, where the symbol appears)
- `col`: Column number (1-based, position of symbol start)

**Options:**
- `--server <name>`: Force specific LSP server
- `--timeout <duration>`: Wait time for LSP response (default: 30s)

## Output Format

```json
[
  { "file": "/path/to/file.go", "line": 42, "col": 5 },
  { "file": "/path/to/other.go", "line": 15, "col": 1 }
]
```

Each result is a location where the symbol is referenced.

## How to Find the Right Position

The LSP query requires the exact position of the symbol you want to look up. Common approaches:

### From error messages
```bash
# If you see "undefined: foo" at line 42, col 10
agentutil lsp refs /path/to/file.go 42 10
```

### From manual inspection
```bash
# Open file, find the line with the symbol
# Count columns from the start of the symbol name (1-based)
```

### Strategy: Start broad, narrow down

If you're unsure which position to query:
1. Run diagnostics to find errors first
2. Look at error locations
3. Query refs at those positions

## Common Patterns

### Find all usages of a function
```bash
agentutil lsp refs src/main.go 42 5  # where 42,5 is the function name
```

### Find definition of a symbol
```bash
# Same command, but refs include the definition location
# Definition is typically at the top of the results
```

### Find all references in a project
```bash
# Get refs for a symbol, then run diagnostics on all files in results
```

### Check if a symbol is used anywhere
```bash
agentutil lsp refs /path/to/file.go <line> <col>
# Empty array = symbol is unused
```

## Technique: Safe Refactoring Workflow

When renaming or moving symbols:

1. **Find all references** — Query the symbol position
2. **Review all locations** — Understand the full scope of changes
3. **Check for conflicts** — Are there multiple symbols with the same name?
4. **Make changes** — Update definition first, then all references
5. **Verify** — Run diagnostics to ensure no new errors

```
┌─────────────────────────────────────────────────────────┐
│                 REFACTORING WORKFLOW                    │
├─────────────────────────────────────────────────────────┤
│                                                         │
│   1. Find symbol position                               │
│      └── Query: agentutil lsp refs <file> <line> <col>  │
│                                                         │
│   2. Review all reference locations                     │
│      └── Check file list, line numbers                  │
│                                                         │
│   3. Verify no naming conflicts                         │
│      └── Are there similar names in scope?              │
│                                                         │
│   4. Update definition                                  │
│      └── Fix the source first                           │
│                                                         │
│   5. Update all references                              │
│      └── Apply same change at each location             │
│                                                         │
│   6. Verify with diagnostics                            │
│      └── agentutil lsp diagnostics <pattern>            │
│                                                         │
└─────────────────────────────────────────────────────────┘
```

## Technique: Bug Investigation

When debugging unexpected behavior:

1. **Find the problematic code** — Look for error locations
2. **Find all usages** — Query refs at that symbol position
3. **Trace the call chain** — Follow references to understand flow
4. **Check for edge cases** — Are there paths not being tested?

## Interpreting Results

**Empty array:** Symbol is never used (dead code candidate)

**Few results:** Symbol is either new, internal, or rarely used

**Many results:** Symbol is widely used — be careful when changing

**Results span multiple files:** Cross-file dependencies to consider

## Common Issues and Solutions

**No references found but symbol exists:**
- LSP might not have finished indexing — try `--timeout 60s`
- Wrong position — verify line and column are correct
- Symbol name might have changed since last analysis

**Too many results:**
- The symbol might be common (e.g., `data`, `error`, `ctx`)
- Try a more specific position — use the actual definition, not a usage

**"No LSP server handles file":**
- File type not supported by installed LSP servers
- LSP server not installed for this language
- File outside workspace root

## Workflow Integration

```
When investigating a bug:
  1. Get diagnostics on the file — surface any errors
  2. Find symbol refs at error location
  3. Trace through the call chain
  4. Find the root cause

When preparing to refactor:
  1. Find all references to the symbol
  2. Understand the full scope
  3. Check for naming conflicts
  4. Make changes definition-first
  5. Verify with diagnostics
```

## Limitations

- Results depend on LSP server accuracy — some servers miss references
- Very large codebases may take time to index
- Some LSP servers don't track certain symbol types (e.g., generated code)

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