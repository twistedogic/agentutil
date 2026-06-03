# agentutil

A CLI toolkit and Go library for AI agent utilities: LSP integration, web fetch, web search, and Wikipedia lookup.

## Features

- **LSP integration**: Auto-detects installed language servers, lazy initialization, diagnostics and references tools
- **Web fetch**: Fetches URLs and converts HTML to clean markdown with extracted links
- **Web search**: Searches the web via DuckDuckGo and returns structured results
- **Wikipedia**: Looks up topics on Wikipedia and returns clean markdown summaries
- **Fantasy agent tools**: All tools expose a `fantasy.AgentTool` for use with [charm.land/fantasy](https://charm.sh)

## Installation

**CLI binary:**

```bash
go install github.com/twistedogic/agentutil@latest
```

**As a library:**

```bash
go get github.com/twistedogic/agentutil
```

## CLI Commands

### `fetch`

Fetch a URL and return its content as clean markdown plus extracted links.

```bash
agentutil fetch <url> [--timeout 30s]
```

Output:
```json
{
  "content": "# Page Title\n\nBody as markdown...",
  "links": ["https://example.com/page"]
}
```

### `search`

Search the web via DuckDuckGo and return structured results.

```bash
agentutil search <query> [--max 10] [-n 10] [--timeout 30s]
```

Output:
```json
{
  "results": [
    {
      "position": 1,
      "title": "The Go Programming Language",
      "url": "https://go.dev/",
      "snippet": "Go is an open source programming language..."
    }
  ]
}
```

> A random 500–2000ms delay is applied before each request to avoid rate limiting.

### `wiki`

Look up a topic on Wikipedia.

```bash
agentutil wiki <topic>
```

### `lsp`

LSP integration tools for AI agents.

```bash
agentutil lsp diagnostics <file>
agentutil lsp references <symbol> [--path <dir>]
agentutil lsp restart [--name <server>]
```

## Architecture

```
agentutil/
├── main.go             # CLI root command
├── fetch.go            # fetch command
├── search.go           # search command
├── wiki.go             # wiki command
├── lsp.go              # lsp command
├── tools/
│   ├── fetch/          # HTML fetch + markdown conversion
│   ├── search/         # DuckDuckGo scraping + result parsing
│   ├── wiki/           # Wikipedia API client
│   └── lsp/            # LSP manager, client, handlers
└── config/             # ServerConfig types
```

## LSP Quick Start

```go
package main

import (
    "context"
    "log/slog"
    "github.com/twistedogic/agentutil/tools/lsp"
)

func main() {
    store := &myConfigStore{}
    manager := lsp.NewManager(store, "/path/to/project")

    manager.SetCallback(func(name string, client *lsp.Client) {
        slog.Info("LSP client event", "name", name, "state", client.GetServerState())
    })

    ctx := context.Background()
    manager.Start(ctx, "/path/to/project/main.go")

    for name, client := range manager.Clients().Seq2() {
        diags := client.GetDiagnostics()
        slog.Info("Diagnostics", "server", name, "count", len(diags))
    }
}
```

## Agent Skills

Pre-built skills for use with AI coding assistants (Crush, Claude Code, etc.) are in `skills/`:

| Skill | Description |
|-------|-------------|
| `web-fetch` | Fetch a URL and read its content |
| `web-search` | Search the web via DuckDuckGo |
| `lsp-diagnostics` | Get LSP diagnostics for a file |
| `lsp-references` | Find all references to a symbol |
| `wiki` | Look up a topic on Wikipedia |

## Credits

LSP integration extracted from [charmbracelet/crush](https://github.com/charmbracelet/crush). Search logic ported from [charmbracelet/crush](https://github.com/charmbracelet/crush/blob/main/internal/agent/tools/search.go). Uses [powernap](https://github.com/charmbracelet/x/tree/main/powernap) for LSP transport.

## License

MIT
