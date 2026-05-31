## Why

The `fetch` tool exists as a `fantasy.AgentTool` for agent runtimes, but agents running via the agentutil CLI have no way to invoke it directly. There is also no skill teaching agents when and how to use URL fetching. This change exposes fetch as a first-class CLI command and adds a `web-fetch` skill.

## What Changes

- Add `agentutil fetch <url>` subcommand that fetches a URL and prints JSON (`content`, `links`) to stdout
- Add `skills/web-fetch/SKILL.md` skill documenting when and how to use `agentutil fetch`

## Capabilities

### New Capabilities

- `cli-fetch`: CLI subcommand `agentutil fetch <url>` wrapping the existing `tools/fetch` package, outputting `FetchResult` JSON

### Modified Capabilities

- `fetch`: Existing fetch spec covers the tool library; no requirement changes needed — CLI usage adds only an invocation layer on top

## Impact

- `cmd/agentutil/`: new `fetch.go` file, wire into `main.go`
- `skills/web-fetch/`: new skill directory and `SKILL.md`
- No changes to `tools/fetch/` (reuse as-is)
- No new dependencies
