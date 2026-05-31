---
name: web-fetch
description: Use when fetching content from a URL, reading a web page, extracting links from a page, or crawling a site. Triggered by requests to "fetch this URL", "read this page", "get content from", "follow links", or any task requiring HTTP content retrieval.
license: MIT
compatibility: Requires agentutil CLI
metadata:
  author: agentutil
  version: "1.0"
---

# Web Fetch

Fetch a URL and get its content as clean markdown plus all extracted links.

## When to Use

- Reading documentation or reference pages
- Extracting structured content from a web page
- Discovering links on a page to crawl further
- Checking what a URL currently returns

## Tool Usage

```bash
agentutil fetch <url>
```

**Arguments:**
- `url`: The URL to fetch (must start with `http://` or `https://`)

**Options:**
- `--timeout <duration>`: HTTP request timeout (default: `30s`)

## Output Format

```json
{
  "content": "# Page Title\n\nPage body as clean markdown...",
  "links": [
    "https://example.com/about",
    "https://example.com/docs",
    "https://other.com/page"
  ]
}
```

- `content`: Page body converted to clean markdown (HTML, script, nav, header, footer stripped)
- `links`: All absolute URLs extracted from `<a href>` elements on the page

## Common Patterns

### Fetch a single page
```bash
agentutil fetch https://pkg.go.dev/net/http
```

### Fetch with extended timeout for slow sites
```bash
agentutil fetch --timeout 60s https://slow-site.example.com
```

### Crawl pattern (fetch → extract links → follow)
```bash
# 1. Fetch the index page
agentutil fetch https://docs.example.com

# 2. Parse links from output, then fetch pages of interest
agentutil fetch https://docs.example.com/getting-started
agentutil fetch https://docs.example.com/api-reference
```

## Edge Cases

| Situation | Behavior |
|-----------|----------|
| Non-HTML response (JSON, plain text, binary) | `content` = raw response body, `links` = `[]` |
| Page > 5MB | Content truncated at 5MB |
| Non-200 HTTP status | Command exits non-zero, error to stderr |
| Invalid URL scheme (ftp://, etc.) | Command exits non-zero, error to stderr |
| Fragment/mailto/javascript hrefs | Filtered out — not included in `links` |

## Installation

```bash
go install github.com/twistedogic/agentutil@latest
```

Verify:
```bash
agentutil fetch --help
```
