## Context

agentutil exposes agent tools via `charm.land/fantasy`. Existing tools (`tools/eino/tools.go`) wrap LSP operations. There is no HTTP or web capability yet. crush's codebase has well-tested fetch helpers we can adapt directly.

## Goals / Non-Goals

**Goals:**
- `FetchTool`: a single `fantasy.AgentTool` that fetches a URL, converts HTML to clean markdown, and extracts absolute links.
- Reusable helpers in `fetch_helpers.go` for HTTP fetch, HTML→markdown, link extraction.

**Non-Goals:**
- Sub-agent orchestration (agentic_fetch style)
- Web search / crawling
- Authentication or cookie handling
- Non-HTML content conversion (PDF, DOCX, etc.)

## Decisions

**Single tool returning both content and links**
Both are derived from the same HTML parse pass. Splitting into two tools would require the caller to fetch twice or pass HTML between tools — unnecessary overhead. A single structured response (`FetchResult{Content, Links}`) is simpler.

**goquery for HTML parsing**
goquery (used in crush) provides a jQuery-like API that makes link extraction and noisy-element removal concise. Stdlib `golang.org/x/net/html` is an alternative but requires more manual tree walking.

**html-to-markdown for content conversion**
Markdown is the most LLM-friendly format. The library is already validated in crush's production use. Plain text extraction loses structure; raw HTML is too noisy.

**Resolve links to absolute URLs**
Raw hrefs (`/page`, `../foo`) are not useful to agents without the base URL context. `url.Parse` + `base.ResolveReference` is stdlib, zero cost. `mailto:`, `javascript:`, and `#fragment`-only hrefs are filtered out.

**Browser User-Agent**
Many sites block non-browser User-Agent strings. Using a realistic UA (as crush does) improves reliability.

## Risks / Trade-offs

- [Large pages] → Content may exceed token limits. Mitigation: document the 5MB read cap; callers can truncate or summarize.
- [goquery dependency] → Adds ~200KB to binary. Acceptable trade-off for reliable HTML parsing.
- [Noisy element removal is heuristic] → Some sites use non-standard structure. Mitigation: the list (script/style/nav/header/footer/aside) covers the common case; edge cases still return parseable content.
