## Context

The `tools/fetch` package already implements URL fetching, HTML-to-markdown conversion, and link extraction as a `fantasy.AgentTool`. The CLI (`cmd/agentutil`) exposes LSP functionality via `agentutil lsp diagnostics` and `agentutil lsp refs`. No CLI surface for fetch exists today, and no skill documents it for agents.

## Goals / Non-Goals

**Goals:**
- Add `agentutil fetch <url>` command that wraps `tools/fetch` and prints `FetchResult` JSON to stdout
- Add `skills/web-fetch/SKILL.md` that teaches agents when and how to use the command

**Non-Goals:**
- Changes to `tools/fetch` internals
- Multiple URL batching
- Output format options (raw HTML, plain text)
- Recursive/crawl orchestration (the skill may describe the pattern, but the tool stays single-URL)

## Decisions

**Reuse `tools/fetch` as-is** — `FetchURLAndConvert` and `FetchResult` are already the right abstraction. The CLI command is a thin wrapper: parse args, call the function, marshal JSON, done. No new logic needed.

**Top-level subcommand, not nested** — `agentutil fetch <url>` rather than `agentutil web fetch <url>`. The LSP commands are grouped because they share flags (`--server`, `--timeout`) and infrastructure (LSP manager). Fetch is standalone; grouping adds no value.

**`--timeout` flag, default 30s** — mirrors the LSP command pattern and the default in `NewFetchTool`. Consistency across CLI surface.

**JSON to stdout, errors to stderr** — same convention as `lsp` commands (`writeJSON`, `writeError`).

## Risks / Trade-offs

- [Large pages] Content could be very large → no mitigation needed at CLI layer; the tool already handles up to 5MB
- [Non-HTML URLs] Raw bytes returned as `content`, empty `links` — documented in skill
