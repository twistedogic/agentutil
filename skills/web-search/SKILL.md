---
name: web-search
description: Use when searching the web, finding information online, looking up a topic, or researching a query. Triggered by requests to "search for", "find information about", "look up", "what is X", or any task requiring web search results.
license: MIT
compatibility: Requires agentutil CLI
metadata:
  author: agentutil
  version: "1.0"
---

# Search

Search the web via DuckDuckGo and get structured results with titles, URLs, and snippets.

## When to Use

- Finding current information or recent events
- Locating documentation, packages, or resources
- Researching a topic before fetching specific pages
- Discovering relevant URLs to follow up with `agentutil fetch`

## Tool Usage

```bash
agentutil search <query>
```

**Arguments:**
- `query`: The search query string

**Options:**
- `--max`, `-n`: Maximum number of results to return (default: `10`)
- `--timeout <duration>`: HTTP request timeout (default: `30s`)

> **Note:** A random 500–2000ms delay is applied before each request to avoid rate limiting.

## Output Format

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

- `results`: Array of search results in ranked order
- `position`: 1-based rank
- `title`: Page title
- `url`: Direct URL (DuckDuckGo redirect URLs are automatically resolved)
- `snippet`: Short summary extracted from the result

## Common Patterns

### Basic search
```bash
agentutil search "golang context package"
```

### Limit results
```bash
agentutil search "openai api pricing" --max 5
```

### Search then fetch (research pattern)
```bash
# 1. Find relevant pages
agentutil search "go embed package tutorial" -n 3

# 2. Fetch the most relevant result
agentutil fetch https://pkg.go.dev/embed
```

## Edge Cases

| Situation | Behavior |
|-----------|----------|
| No results found | `{"results": []}`, exits with code 0 |
| DDG returns non-200 | Command exits non-zero, error to stderr |
| Timeout exceeded | Command exits non-zero with context deadline error |
| Query is empty | Command exits non-zero with usage error |

## Installation

```bash
go install github.com/twistedogic/agentutil@latest
```

Verify:
```bash
agentutil search --help
```
